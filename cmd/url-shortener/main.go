package main

import (
	"httpServer_project/internal/config"
	"httpServer_project/internal/storage/postgres"
	"httpServer_project/lib/logger/slg"
	"log/slog"
	"os"
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

	storage, err := postgres.New(cfg.Postgres.Dsn)
	if err != nil {
		log.Error("failed to initialize storage", slg.Err(err))
		os.Exit(1)
	}

	defer storage.Close() //// Закрываем пул соединений при завершении программы

	_ = storage

	// TODO: init router: chi, "chi render"

	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
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
