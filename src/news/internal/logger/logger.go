// Package logger предоставляет функции для инициализации и использования логгера в приложении.
package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Init(env string) {
	switch env {
	case envLocal:
		Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
