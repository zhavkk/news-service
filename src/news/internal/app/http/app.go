package httpapp

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/zhavkk/news-service/src/news/internal/config"
	v1 "github.com/zhavkk/news-service/src/news/internal/handlers/v1"
	"github.com/zhavkk/news-service/src/news/internal/logger"
)

type HTTPApp struct {
	fiberApp *fiber.App
	port     int
}

func New(cfg *config.Config, newsService v1.NewsService) *HTTPApp {

	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		AppName:      "News Service",
	})

	setupMiddlewares(app)

	setupRoutes(app, newsService)
	return &HTTPApp{
		fiberApp: app,
		port:     cfg.HTTP.Port,
	}
}

func (a *HTTPApp) Start() error {
	logger.Log.Info("Starting HTTP server", "port", a.port)
	return a.fiberApp.Listen(fmt.Sprintf(":%d", a.port))
}

func (a *HTTPApp) Stop(ctx context.Context) error {
	logger.Log.Info("Stopping HTTP server", "port", a.port)
	return a.fiberApp.Shutdown()
}

func (a *HTTPApp) App() *fiber.App {
	return a.fiberApp
}

func setupMiddlewares(app *fiber.App) {
	app.Use(requestid.New())

	// cors and etc

}

func setupRoutes(app *fiber.App, newsService v1.NewsService) {
	api := app.Group("/api")
	v1Group := api.Group("/v1")

	newsHandler := v1.NewHandler(newsService)
	newsHandler.RegisterRoutes(v1Group)
}
