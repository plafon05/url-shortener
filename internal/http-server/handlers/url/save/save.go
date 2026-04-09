package save

import (
	"context"
	"errors"
	"net/http"

	"log/slog"

	"httpServer_project/internal/storage"
	"httpServer_project/lib/aliasgen"
	resp "httpServer_project/lib/api/response"
	"httpServer_project/lib/logger/slg"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

// URLSaver — интерфейс для сохранения URL с алиасом.
type URLSaver interface {
	SaveURL(ctx context.Context, urlToSave, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	// Инициализируем валидатор один раз, а не при каждом запросе.
	validate := validator.New()

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.handlers.url.save.New"

		// Используем := чтобы не мутировать внешний логгер.
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request", slg.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validate.Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", slg.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = aliasgen.GenerateAlias(req.URL)
			log.Info("alias generated", slog.String("alias", alias))
		}

		id, err := urlSaver.SaveURL(r.Context(), req.URL, alias)
		if errors.Is(err, storage.ErrAliasExists) {
			log.Info("alias already exists", slog.String("alias", alias))
			render.Status(r, http.StatusConflict) // 409
			render.JSON(w, r, resp.Error("alias already exists"))
			return
		}
		if err != nil {
			log.Error("failed to save URL", slg.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("URL saved", slog.Int64("id", id))
		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
