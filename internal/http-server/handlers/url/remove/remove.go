package remove

import (
	"context"
	"errors"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"httpServer_project/internal/storage"
	resp "httpServer_project/lib/api/response"
	"httpServer_project/lib/logger/slg"
)

// URLDeleter это интерфейс для удаления URL по алиасу.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(ctx context.Context, alias string) error
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias пустой")

			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		if err := deleter.DeleteURL(r.Context(), alias); err != nil {
			// Алиас не найден
			if errors.Is(err, storage.ErrAliasNotFound) {
				log.Info("url не найден", "alias", alias)
				render.JSON(w, r, resp.Error("not found"))
				return
			}

			// Любая другая ошибка
			log.Error("ошибка при удалении url", slg.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("url по алиасу удален", slog.String("alias", alias))

		render.JSON(w, r, resp.OK())
	}
}
