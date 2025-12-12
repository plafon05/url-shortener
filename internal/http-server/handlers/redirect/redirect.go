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

// URLGetter это интерфейс для получения URL по алиасу.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
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

		// Получение алиаса из URL-параметров
		// Извлекаем алиас
		alias := chi.URLParam(r, "alias")

		if alias == "" {
			// Логируем с контекстом
			log.Info("пустой алиас в запросе",
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.String("client_ip", r.RemoteAddr),
				slog.String("user_agent", r.Header.Get("User-Agent")),
			)

			render.Status(r, http.StatusBadRequest) // 400

			render.JSON(w, r, resp.Error("алиас не может быть пустым"))

			return
		}

		resURL, err := urlGetter.GetURL(r.Context(), alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url не найден", "alias", alias)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("ошибка при получении url", slg.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("получен url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
