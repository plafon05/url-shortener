package save

import (
	resp "httpServer_project/lib/api/response"
	"log/slog"
	"net/http"

	"httpServer_project/lib/logger/slg"

	"httpServer_project/lib/aliasgen"

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

type URLSever interface {
	SeveURL(urlToSave, alias string) (int64, error)
}

func New(log *slog.Logger, urlSever URLSever) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("не удалось расшифровать запрос", slg.Err(err))
			render.JSON(w, r, resp.Error("не удалось расшифровать запрос"))
			return
		}

		log.Info("тело запроса расшифровано", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("недопустимый запрос", slg.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = aliasgen.GenerateAlias(req.URL)
			log.Info("сгенерирован алиас", slog.String("alias", alias))
		}
	}
}
