package apierrors

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/vladgrskkh/todo/pkg/jsonhttp"
)

// logError logs an error with the request method, URL, and stack trace.
func logError(logger *slog.Logger, r *http.Request, err error) {
	logger.Error(err.Error(),
		slog.String("request_method", r.Method),
		slog.String("request_url", r.URL.String()),
		slog.String("trace", string(debug.Stack())))
}

// errorResponse writes a JSON response with a provided status code and message
// to the http.ResponseWriter.
func errorResponse(logger *slog.Logger, w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	data := jsonhttp.Envelope{
		"error": message,
	}

	err := jsonhttp.WriteJSON(w, status, data, nil)
	if err != nil {
		logError(logger, r, err)
		w.WriteHeader(500)
	}
}

func BadRequestResponse(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(logger, w, r, http.StatusBadRequest, err.Error())
}

func ServerErrorResponse(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	logError(logger, r, err)

	message := "server encountered a problem and could not process your request"
	errorResponse(logger, w, r, http.StatusInternalServerError, message)
}

func NotFoundResponse(logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	message := "requested resource could not be found"
	errorResponse(logger, w, r, http.StatusNotFound, message)
}
