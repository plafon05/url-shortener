package main

import (
	"httpServer_project/internal/config"
	"httpServer_project/internal/http-server/handlers/url/save"
	mwLogger "httpServer_project/internal/http-server/middleware/logger"
	"httpServer_project/internal/storage/postgres"
	"httpServer_project/lib/logger/handlers/slogpretty"
	"httpServer_project/lib/logger/slg"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MastLoad()

	log := setupLogger(cfg.Env)
	log.Info("start url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug massages are enabled")

	// Инициализирую хранилище моих данных
	storage, err := postgres.New(cfg.Postgres.Dsn)

	if err != nil {
		log.Error("failed to initialize storage", slg.Err(err))
		os.Exit(1)
	}

	defer storage.Close() // Закрываю пул соединений при завершении программы

	// Иницилизация роутера на go-chi
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, storage))

	// TODO: run server
}

// Настройка логгера в зависимости от окружения
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

// Настройка pretty slog логгера
func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
