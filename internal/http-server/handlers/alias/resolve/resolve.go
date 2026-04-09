package resolve

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"httpServer_project/internal/storage"
	resp "httpServer_project/lib/api/response"
	"httpServer_project/lib/logger/slg"
)

// AliasGetter — интерфейс для получения алиасов по исходному URL.
type AliasGetter interface {
	GetAliasesByURL(ctx context.Context, url string) ([]string, error)
}

type Response struct {
	resp.Response
	Alias []string `json:"alias,omitempty"`
}

func New(log *slog.Logger, aliasGetter AliasGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.alias.resolve.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		urlParam := r.URL.Query().Get("url")
		if urlParam == "" {
			log.Info("empty url in request",
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.String("client_ip", r.RemoteAddr),
				slog.String("user_agent", r.Header.Get("User-Agent")),
			)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("url cannot be empty"))
			return
		}

		// Валидация URL через стандартную библиотеку вместо strings.HasPrefix.
		if parsed, err := url.ParseRequestURI(urlParam); err != nil ||
			(parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" {
			log.Info("invalid URL", slog.String("url", urlParam))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid URL"))
			return
		}

		aliases, err := aliasGetter.GetAliasesByURL(r.Context(), urlParam)
		if errors.Is(err, storage.ErrAliasNotFound) {
			log.Info("aliases not found", slog.String("url", urlParam))
			render.Status(r, http.StatusNotFound) // 404
			render.JSON(w, r, resp.Error("not found"))
			return
		}
		if err != nil {
			log.Error("failed to get aliases",
				slg.Err(err),
				slog.String("url", urlParam),
			)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("aliases retrieved",
			slog.Any("aliases", aliases),
			slog.String("url", urlParam),
		)

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    aliases,
		})
	}
}
