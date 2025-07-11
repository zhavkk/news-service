package service

import (
	"context"

	"github.com/zhavkk/news-service/src/news/internal/models"
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
		sortDir string) ([]*models.News, int64, error)
}

type NewsService struct {
	repo NewsRepository
}
