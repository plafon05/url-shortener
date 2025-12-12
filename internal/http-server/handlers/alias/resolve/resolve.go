package resolve

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"httpServer_project/internal/storage"
	resp "httpServer_project/lib/api/response"
	"httpServer_project/lib/logger/slg"
)

type AliasGetter interface {
	GetAliasByURL(ctx context.Context, url string) (string, error)
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

func New(log *slog.Logger, aliasGetter AliasGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.alias.resolve.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Получение url из query-параметров
		urlParam := r.URL.Query().Get("url")

		if urlParam == "" {
			// Логируем с контекстом
			log.Info("пустой url в запросе",
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.String("client_ip", r.RemoteAddr),
				slog.String("user_agent", r.Header.Get("User-Agent")),
			)

			render.Status(r, http.StatusBadRequest) // 400
			render.JSON(w, r, resp.Error("url не может быть пустым"))
			return
		}

		// Базовая валидация URL
		if !strings.HasPrefix(urlParam, "http://") &&
			!strings.HasPrefix(urlParam, "https://") {
			log.Info("невалидный URL", slog.String("url", urlParam))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("URL должен начинаться с http:// или https://"))
			return
		}

		alias, err := aliasGetter.GetAliasByURL(r.Context(), urlParam)

		if errors.Is(err, storage.ErrAliasNotFound) {
			log.Info("alias не найден", "url", urlParam)
			render.JSON(w, r, resp.Error("not found"))
			return
		}

		if err != nil {
			log.Error("ошибка при получении alias",
				slg.Err(err),
				slog.String("url", urlParam),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("получен alias",
			slog.String("alias", alias),
			slog.String("url", urlParam),
		)

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
