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

// URLDeleter — интерфейс для удаления URL по алиасу.
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
			log.Info("empty alias",
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
			)
			render.Status(r, http.StatusBadRequest) // 400
			render.JSON(w, r, resp.Error("alias cannot be empty"))
			return
		}

		if err := deleter.DeleteURL(r.Context(), alias); err != nil {
			if errors.Is(err, storage.ErrAliasNotFound) {
				log.Info("alias not found", slog.String("alias", alias))
				render.Status(r, http.StatusNotFound) // 404
				render.JSON(w, r, resp.Error("not found"))
				return
			}

			log.Error("failed to delete url",
				slg.Err(err),
				slog.String("alias", alias),
			)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("url deleted by alias", slog.String("alias", alias))
		render.JSON(w, r, resp.OK())
	}
}
