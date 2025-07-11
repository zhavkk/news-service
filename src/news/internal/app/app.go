package app

import (
	httpapp "github.com/zhavkk/news-service/src/news/internal/app/http"
	"github.com/zhavkk/news-service/src/news/internal/config"
)

type App struct {
	HTTPServer *httpapp.HTTPApp
}

func NewApp(cfg *config.Config) *App {
	httpServer := httpapp.New(cfg)

	return &App{
		HTTPServer: httpServer,
	}
}
