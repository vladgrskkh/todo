package handlers

import (
	"log/slog"
	"net/http"

	"github.com/vladgrskkh/todo/internal/apierrors"
	"github.com/vladgrskkh/todo/pkg/jsonhttp"
)

func NewHealthCheckHandler(logger *slog.Logger, env, version string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := jsonhttp.Envelope{
			"status":  "avaliable",
			"env":     env,
			"version": version,
		}

		err := jsonhttp.WriteJSON(w, http.StatusOK, data, nil)
		if err != nil {
			apierrors.ServerErrorResponse(logger, w, r, err)
		}
	})
}
