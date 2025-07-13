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

// CreateNews godoc
// @Summary      Create a news item
// @Description  Adds a new news item to the database with content blocks
// @Tags         news
// @Accept       json
// @Produce      json
// @Param        news  body      dto.CreateNewsRequest  true  "News to create"
// @Success      201   {object}  dto.NewsResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /news [post]
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

// GetNewsByID godoc
// @Summary      Get a news item by ID
// @Description  Retrieves a news item and its content blocks by its ID
// @Tags         news
// @Produce      json
// @Param        id                path      string  true  "News ID"
// @Param        check_visibility  query     bool    false "Check visibility (start/end time)" default(true)
// @Success      200               {object}  dto.NewsResponse
// @Failure      400               {object}  dto.ErrorResponse
// @Failure      404               {object}  dto.ErrorResponse
// @Failure      500               {object}  dto.ErrorResponse
// @Router       /news/{id} [get]
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

// ListNews godoc
// @Summary      Get a list of news
// @Description  Retrieves a list of news items with pagination, filtering, and sorting
// @Tags         news
// @Produce      json
// @Param        page              query     int     false "Page number for pagination" default(1)
// @Param        limit             query     int     false "Number of items per page" default(10)
// @Param        search            query     string  false "Search term for news titles"
// @Param        category          query     string  false "Filter by category"
// @Param        sort_by           query     string  false "Field to sort by" Enums(created_at, start_time, end_time, title, category) default(created_at)
// @Param        sort_dir          query     string  false "Sort direction" Enums(asc, desc) default(desc)
// @Param        check_visibility  query     bool    false "Check visibility (start/end time)" default(true)
// @Success      200               {object}  dto.NewsListResponse
// @Failure      400               {object}  dto.ErrorResponse
// @Failure      500               {object}  dto.ErrorResponse
// @Router       /news [get]
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

// UpdateNews godoc
// @Summary      Update a news item
// @Description  Updates a news item's details and content blocks by its ID
// @Tags         news
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "News ID"
// @Param        news  body      dto.UpdateNewsRequest  true  "Fields to update"
// @Success      200   {object}  dto.UpdateNewsResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /news/{id} [put]
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

// DeleteNews godoc
// @Summary      Delete a news item
// @Description  Deletes a news item by its ID
// @Tags         news
// @Produce      json
// @Param        id   path      string  true  "News ID"
// @Success      200  {object}  dto.DeleteNewsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /news/{id} [delete]
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
