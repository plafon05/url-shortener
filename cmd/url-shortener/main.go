package main

import (
	"httpServer_project/internal/config"
	"httpServer_project/internal/http-server/handlers/alias/resolve"
	"httpServer_project/internal/http-server/handlers/redirect"
	"httpServer_project/internal/http-server/handlers/url/remove"
	"httpServer_project/internal/http-server/handlers/url/save"
	mwLogger "httpServer_project/internal/http-server/middleware/logger"
	"httpServer_project/internal/storage/postgres"
	"httpServer_project/lib/logger/handlers/slogpretty"
	"httpServer_project/lib/logger/slg"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("запуск url-shortener", slog.String("env", cfg.Env))
	log.Debug("отладочные сообщения включены")

	// ---------------------------
	// Подключение к Postgres с retry
	var storage *postgres.Storage
	var err error

	for i := 1; i <= 10; i++ {
		storage, err = postgres.New(cfg.Postgres.Dsn)
		if err == nil {
			break
		}

		log.Warn(
			"postgres недоступен, повторная попытка",
			slog.Int("attempt", i),
			slog.Any("error", err),
		)

		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Error("postgres так и не стал доступен", slg.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	// ---------------------------
	// Инициализация роутера
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", remove.New(log, storage))
		r.Get("/aliases", resolve.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("сервер запущен", slog.String("address", cfg.HTTPServer.Address))

	svr := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := svr.ListenAndServe(); err != nil {
		log.Error("ошибка при запуске сервера", slg.Err(err))
		os.Exit(1)
	}

	log.Error("сервер остановлен")
}

// ---------------------------
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

// pretty slog для локальной разработки
func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}
