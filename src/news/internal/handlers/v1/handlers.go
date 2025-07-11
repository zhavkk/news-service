package v1

import (
	"context"

	"github.com/zhavkk/news-service/src/news/internal/dto"
)

type NewsService interface {
	CreateNews(ctx context.Context, req dto.CreateNewsRequest) (*dto.NewsResponse, error)
	UpdateNews(ctx context.Context, req dto.UpdateNewsRequest) (*dto.UpdateNewsResponse, error)
	GetNewsByID(ctx context.Context, req dto.GetNewsByIDRequest) (*dto.NewsResponse, error)
	DeleteNews(ctx context.Context, req dto.DeleteNewsRequest) (*dto.DeleteNewsResponse, error)
	ListNews(ctx context.Context, req dto.NewsListRequest) (*dto.NewsListResponse, error)
}

type NewsHandler struct {
	newsService NewsService
}

func NewHandler(newsService NewsService) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
	}
}
