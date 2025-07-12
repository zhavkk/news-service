package app

import (
	"context"

	httpapp "github.com/zhavkk/news-service/src/news/internal/app/http"
	"github.com/zhavkk/news-service/src/news/internal/config"
	"github.com/zhavkk/news-service/src/news/internal/logger"
	"github.com/zhavkk/news-service/src/news/internal/repository/postgres"
	"github.com/zhavkk/news-service/src/news/internal/service"
	"github.com/zhavkk/news-service/src/news/internal/storage"
)

type App struct {
	HTTPServer *httpapp.HTTPApp
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	logger.Log.Info("Initializing application with config", "env", cfg.Env, "port", cfg.HTTP.Port)
	txManager, err := storage.NewTxManager(ctx, cfg)
	if err != nil {
		return nil, err
	}

	newsRepo := postgres.NewNewsRepository(txManager.GetDatabase())

	redis, err := storage.NewRedisClient(ctx, &cfg.Redis)
	if err != nil {
		logger.Log.Error("Failed to initialize Redis client", "error", err)
		return nil, err
	}

	newsService := service.NewNewsService(newsRepo, txManager, redis, cfg.Redis.CacheTTL)

	httpServer := httpapp.New(cfg, newsService)
	logger.Log.Info("Application initialized successfully", "env", cfg.Env, "port", cfg.HTTP.Port)

	return &App{
		HTTPServer: httpServer,
	}, nil
}
