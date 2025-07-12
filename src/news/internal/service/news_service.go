package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zhavkk/news-service/src/news/internal/dto"
	"github.com/zhavkk/news-service/src/news/internal/logger"
	"github.com/zhavkk/news-service/src/news/internal/models"
	"github.com/zhavkk/news-service/src/news/internal/repository/postgres"
	"github.com/zhavkk/news-service/src/news/internal/storage"
)

type NewsRepository interface {
	Create(
		ctx context.Context,
		news *models.News,
	) error

	GetByID(
		ctx context.Context,
		id int64,
	) (*models.News, error)

	Update(
		ctx context.Context,
		news *models.News,
	) error

	Delete(
		ctx context.Context,
		id int64,
	) error

	List(
		ctx context.Context,
		offset int,
		limit int,
		search string,
		category string,
		sortBy string,
		sortDir string,
		checkVisibility bool,
	) ([]*models.News, int64, error)
}

type RedisClient interface {
	GetRedis() *redis.Client
}

type NewsService struct {
	newsRepo  NewsRepository
	txManager storage.TxManagerInterface
	redis     RedisClient
	cacheTTL  time.Duration
}

func NewNewsService(
	newsRepo NewsRepository,
	txManager storage.TxManagerInterface,
	redis RedisClient,
	cacheTTL time.Duration,
) *NewsService {
	return &NewsService{
		newsRepo:  newsRepo,
		txManager: txManager,
		redis:     redis,
		cacheTTL:  cacheTTL,
	}
}

func (s *NewsService) CreateNews(
	ctx context.Context,
	req dto.CreateNewsRequest,
) (*dto.NewsResponse, error) {
	const op = "service.NewsService.CreateNews"

	var resp *dto.NewsResponse

	err := s.txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		news := &models.News{
			Title:     req.Title,
			Category:  req.Category,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
		}

		for _, block := range req.Content {
			contentBlock := models.ContentBlock{
				Type:     models.BlockType(block.Type),
				Content:  block.Content,
				Position: block.Position,
			}
			news.Content = append(news.Content, contentBlock)
		}

		if err := s.newsRepo.Create(ctx, news); err != nil {
			return err
		}

		contentDTO := make([]dto.ContentBlockResponse, len(news.Content))
		for i, block := range news.Content {
			contentDTO[i] = dto.ContentBlockResponse{
				ID:       strconv.FormatInt(block.ID, 10),
				Type:     string(block.Type),
				Content:  block.Content,
				Position: block.Position,
			}
		}
		logger.Log.Info(op, "News created successfully", news.ID)

		resp = &dto.NewsResponse{
			ID:        strconv.FormatInt(news.ID, 10),
			Title:     news.Title,
			Category:  news.Category,
			StartTime: news.StartTime,
			EndTime:   news.EndTime,
			Content:   contentDTO,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func (s *NewsService) GetNewsByID(
	ctx context.Context,
	req dto.GetNewsByIDRequest,
) (*dto.NewsResponse, error) {
	const op = "service.NewsService.GetNewsByID"

	cacheKey := fmt.Sprintf("news:%s", req.ID)

	cachedNews, err := s.redis.GetRedis().Get(ctx, cacheKey).Result()
	if err == nil {
		logger.Log.Info(op, "Cache hit for news ID", req.ID)
		var newsResp dto.NewsResponse
		if err := json.Unmarshal([]byte(cachedNews), &newsResp); err != nil {
			logger.Log.Error(op, "Failed to unmarshal cached news", err)
			return nil, err
		}
		return &newsResp, nil
	}

	newsID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		logger.Log.Error(op, "Failed to parse news ID", err)
		return nil, err
	}

	var resp *dto.NewsResponse

	err = s.txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		news, err := s.newsRepo.GetByID(ctx, newsID)
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				return postgres.ErrNotFound
			}
			return err
		}

		if req.CheckVisibility && !news.IsVisible() {
			return postgres.ErrNotFound
		}

		contentDTO := make([]dto.ContentBlockResponse, len(news.Content))
		for i, block := range news.Content {
			contentDTO[i] = dto.ContentBlockResponse{
				ID:       strconv.FormatInt(block.ID, 10),
				Type:     string(block.Type),
				Content:  block.Content,
				Position: block.Position,
			}
		}
		logger.Log.Info(op, "News retrieved successfully", news.ID)

		resp = &dto.NewsResponse{
			ID:        strconv.FormatInt(news.ID, 10),
			Title:     news.Title,
			Category:  news.Category,
			CreatedAt: news.CreatedAt,
			StartTime: news.StartTime,
			EndTime:   news.EndTime,
			Content:   contentDTO,
		}

		toCache, err := json.Marshal(resp)
		if err != nil {
			logger.Log.Error(op, "Failed to marshal news for caching", err)
		} else {
			if err := s.redis.GetRedis().Set(ctx, cacheKey, toCache, s.cacheTTL).Err(); err != nil {
				logger.Log.Error(op, "Failed to set cache", cacheKey, "error", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *NewsService) ListNews(
	ctx context.Context,
	req dto.NewsListRequest,
) (*dto.NewsListResponse, error) {
	const op = "service.NewsService.ListNews"

	offset := (req.Page - 1) * req.Limit

	var resp *dto.NewsListResponse

	err := s.txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		newsList, totalCount, err := s.newsRepo.List(
			ctx,
			offset,
			req.Limit,
			req.Search,
			req.Category,
			req.SortBy,
			req.SortDir,
			req.CheckVisibility,
		)
		if err != nil {
			return err
		}

		items := make([]dto.NewsResponse, 0, len(newsList))

		for _, news := range newsList {
			if req.CheckVisibility && !news.IsVisible() {
				continue
			}

			contentDTO := make([]dto.ContentBlockResponse, len(news.Content))
			for i, block := range news.Content {
				contentDTO[i] = dto.ContentBlockResponse{
					ID:       strconv.FormatInt(block.ID, 10),
					Type:     string(block.Type),
					Content:  block.Content,
					Position: block.Position,
				}
			}

			items = append(items, dto.NewsResponse{
				ID:        strconv.FormatInt(news.ID, 10),
				Title:     news.Title,
				Category:  news.Category,
				CreatedAt: news.CreatedAt,
				StartTime: news.StartTime,
				EndTime:   news.EndTime,
				Content:   contentDTO,
			})
		}
		logger.Log.Info(op, "News list retrieved successfully, total count: ", totalCount)
		resp = &dto.NewsListResponse{
			Items:      items,
			TotalCount: totalCount,
			Page:       req.Page,
			Limit:      req.Limit,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *NewsService) UpdateNews(
	ctx context.Context,
	req dto.UpdateNewsRequest,
) (*dto.UpdateNewsResponse, error) {
	const op = "service.NewsService.UpdateNews"

	newsID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		logger.Log.Error(op, "Failed to parse news ID", err)
		return nil, err
	}

	var resp *dto.UpdateNewsResponse

	err = s.txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		news, err := s.newsRepo.GetByID(ctx, newsID)
		if err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				return postgres.ErrNotFound
			}
			return err
		}

		if req.Title != "" {
			news.Title = req.Title
		}
		if req.Category != "" {
			news.Category = req.Category
		}
		if req.StartTime != nil {
			news.StartTime = *req.StartTime
		}
		if req.EndTime != nil {
			news.EndTime = *req.EndTime
		}

		if len(req.Content) > 0 {
			news.Content = make([]models.ContentBlock, len(req.Content))
			for i, block := range req.Content {
				contentBlock := models.ContentBlock{
					Type:     models.BlockType(block.Type),
					Content:  block.Content,
					Position: block.Position,
				}
				news.Content[i] = contentBlock
			}
		}

		if err := s.newsRepo.Update(ctx, news); err != nil {
			return err
		}

		logger.Log.Info(op, "News updated successfully", news.ID)

		resp = &dto.UpdateNewsResponse{
			ID:        strconv.FormatInt(news.ID, 10),
			UpdatedAt: time.Now(),
			Message:   "News updated successfully",
		}

		return nil

	})
	if err != nil {
		return nil, err
	}
	logger.Log.Info(op, "News updated successfully", req.ID)

	cacheKey := fmt.Sprintf("news:%s", req.ID)
	if err := s.redis.GetRedis().Del(ctx, cacheKey).Err(); err != nil {
		logger.Log.Error(op, "Failed to invalidate cache", cacheKey, "error", err)
	}

	return resp, nil
}

func (s *NewsService) DeleteNews(
	ctx context.Context,
	req dto.DeleteNewsRequest,
) (*dto.DeleteNewsResponse, error) {
	const op = "service.NewsService.DeleteNews"

	newsID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		logger.Log.Error(op, "Failed to parse news ID", err)
		return nil, err
	}

	var resp *dto.DeleteNewsResponse

	err = s.txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		if err := s.newsRepo.Delete(ctx, newsID); err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				return postgres.ErrNotFound
			}
			return err
		}

		logger.Log.Info(op, "News deleted successfully", newsID)

		resp = &dto.DeleteNewsResponse{
			ID:      req.ID,
			Message: "News deleted successfully",
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("news:%s", req.ID)
	if err := s.redis.GetRedis().Del(ctx, cacheKey).Err(); err != nil {
		logger.Log.Error(op, "Failed to invalidate cache", cacheKey, "error", err)
	}

	return resp, nil
}
