package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhavkk/news-service/src/news/internal/config"
	"github.com/zhavkk/news-service/src/news/internal/logger"
	"github.com/zhavkk/news-service/src/news/internal/models"
	postgres "github.com/zhavkk/news-service/src/news/internal/repository/postgres"
	"github.com/zhavkk/news-service/src/news/internal/storage"
)

func setupTestDB(t *testing.T) (*postgres.NewsRepository, storage.TxManagerInterface, func()) {

	dbURL := os.Getenv("DB_URL_TEST")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/news_service_test?sslmode=disable"
	}

	cfg := &config.Config{
		DBURL: dbURL,
	}
	logger.Init("local")
	db, err := storage.NewStorage(context.Background(), cfg)
	require.NoError(t, err, "Failed to connect to test database")

	cleanup := func() {
		_, err := db.GetPool().Exec(context.Background(), "TRUNCATE TABLE news, content_blocks RESTART IDENTITY CASCADE")
		require.NoError(t, err)
		require.NoError(t, db.Close())

	}

	repo := postgres.NewNewsRepository(db)
	txManager := storage.NewTxManagerForTest(db)

	return repo, txManager, cleanup
}

func TestNewsRepository_CreateAndGet(t *testing.T) {
	repo, txManager, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now().Add(time.Hour)
	newsToCreate := &models.News{
		Title:     "Test Create News",
		Category:  "Testing",
		StartTime: startTime,
		EndTime:   endTime,
		Content: []models.ContentBlock{
			{Type: "text", Content: "First block", Position: 1},
			{Type: "link", Content: "http://example.com", Position: 2},
		},
	}

	err := txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		return repo.Create(ctx, newsToCreate)
	})
	require.NoError(t, err)
	require.NotZero(t, newsToCreate.ID)

	retrievedNews, err := repo.GetByID(ctx, newsToCreate.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedNews)

	assert.Equal(t, "Test Create News", retrievedNews.Title)
	assert.Equal(t, "Testing", retrievedNews.Category)
	assert.Len(t, retrievedNews.Content, 2)
	assert.Equal(t, "First block", retrievedNews.Content[0].Content)
}

func TestNewsRepository_Update(t *testing.T) {
	repo, txManager, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	news := &models.News{Title: "Initial Title", Category: "Initial", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)}
	err := txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		return repo.Create(ctx, news)
	})
	require.NoError(t, err)

	news.Title = "Updated Title"
	news.Content = []models.ContentBlock{{Type: "text", Content: "Updated content", Position: 1}}

	err = txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		return repo.Update(ctx, news)
	})
	require.NoError(t, err)

	updatedNews, err := repo.GetByID(ctx, news.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedNews.Title)
	assert.Len(t, updatedNews.Content, 1)
	assert.Equal(t, "Updated content", updatedNews.Content[0].Content)
}

func TestNewsRepository_Delete(t *testing.T) {
	repo, txManager, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	news := &models.News{Title: "To Be Deleted", Category: "Temp", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)}
	err := txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		return repo.Create(ctx, news)
	})
	require.NoError(t, err)

	err = txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		return repo.Delete(ctx, news.ID)
	})
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, news.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound)
}

func TestNewsRepository_List(t *testing.T) {
	repo, txManager, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	news1 := &models.News{Title: "Sport News", Category: "Sport", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)}
	news2 := &models.News{Title: "Finance News", Category: "Finance", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)}
	err := txManager.RunReadCommited(ctx, func(ctx context.Context) error {
		if err := repo.Create(ctx, news1); err != nil {
			return err
		}
		return repo.Create(ctx, news2)
	})
	require.NoError(t, err)

	newsList, totalCount, err := repo.List(ctx, 0, 10, "", "", "created_at", "desc", true)
	require.NoError(t, err)
	assert.Equal(t, int64(2), totalCount)
	assert.Len(t, newsList, 2)

	newsList, totalCount, err = repo.List(ctx, 0, 10, "", "Sport", "created_at", "desc", true)
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalCount)
	assert.Len(t, newsList, 1)
	assert.Equal(t, "Sport News", newsList[0].Title)
}
