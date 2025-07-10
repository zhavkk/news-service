package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhavkk/news-service/src/news/internal/models"
)

type NewsRepository struct {
	db *pgxpool.Pool
}

func NewNewsRepository(db *pgxpool.Pool) *NewsRepository {
	return &NewsRepository{
		db: db,
	}
}

func (r *NewsRepository) Create(ctx context.Context, news *models.News) error {

}
