package redirect

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

// URLGetter — интерфейс для получения URL по алиасу.
type URLGetter interface {
	GetURL(ctx context.Context, alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("empty alias",
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.String("client_ip", r.RemoteAddr),
				slog.String("user_agent", r.Header.Get("User-Agent")),
			)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("alias cannot be empty"))
			return
		}

		resURL, err := urlGetter.GetURL(r.Context(), alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.String("alias", alias))
			render.Status(r, http.StatusNotFound) // 404
			render.JSON(w, r, resp.Error("not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url",
				slg.Err(err),
				slog.String("alias", alias),
			)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("url retrieved", slog.String("url", resURL))
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
