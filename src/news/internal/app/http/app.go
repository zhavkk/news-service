package httpapp

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/zhavkk/news-service/src/news/internal/config"
	"github.com/zhavkk/news-service/src/news/internal/logger"
)

type HTTPApp struct {
	fiberApp *fiber.App
	port     int
}

func New(cfg *config.Config) *HTTPApp {

	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		AppName:      "News Service",
	})

	setupMiddlewares(app)

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
