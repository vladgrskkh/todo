package routes

import (
	"log/slog"
	"net/http"

	"github.com/vladgrskkh/todo/internal/handlers"
	"github.com/vladgrskkh/todo/internal/handlers/middleware"
	"github.com/vladgrskkh/todo/internal/service"
)

func Routes(logger *slog.Logger, service *service.TodoService) http.Handler {
	router := http.NewServeMux()

	// middleware init
	requestLogger := middleware.RequestLogger(logger)
	recoverPanic := middleware.RecoverPanic(logger)

	router.HandleFunc("GET /todos/{id}", handlers.NewGetTaskHandler(logger, service))
	router.HandleFunc("GET /todos", handlers.NewGetAllTasksHandler(logger, service))
	router.HandleFunc("POST /todos", handlers.NewPostTaskHandler(logger, service))
	router.HandleFunc("PUT /todos/{id}", handlers.NewTaskUpdater(logger, service))
	router.HandleFunc("DELETE /todos/{id}", handlers.NewDeleteTaskHandler(logger, service))

	return requestLogger(recoverPanic(router))
}
