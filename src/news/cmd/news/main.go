package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/zhavkk/news-service/src/news/internal/app"
	"github.com/zhavkk/news-service/src/news/internal/config"
	"github.com/zhavkk/news-service/src/news/internal/logger"
)

// @title           News Service API
// @version         1.0
// @description     service for creating and managing news.
// @termsOfService  http://swagger.io/terms/

// @host      localhost:8080
// @BasePath  /api/v1
func main() {

	cfg := config.MustLoad("src/news/config/config.yml")

	logger.Init(cfg.Env)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		logger.Log.Error("Failed to initialize application", "error", err)
		return
	}
	go func() {
		if err := application.HTTPServer.Start(); err != nil {
			logger.Log.Error("Failed to start application", "error", err)
			os.Exit(1)
		}
	}()
	logger.Log.Info("Application started successfully", "env", cfg.Env, "port", cfg.HTTP.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := application.HTTPServer.Stop(ctx); err != nil {
		logger.Log.Error("Failed to stop application gracefully", "error", err)
		os.Exit(1)
	}
	logger.Log.Info("Application stopped gracefully", "env", cfg.Env, "port", cfg.HTTP.Port)

}
