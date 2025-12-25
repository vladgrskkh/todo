package routes

import (
	"log/slog"
	"net/http"

	"github.com/vladgrskkh/todo/internal/handlers"
	"github.com/vladgrskkh/todo/internal/service"
)

func Routes(logger *slog.Logger, service *service.TodoService) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /v1/tasks/{id}", handlers.NewGetTaskHandler(logger, service))
	router.HandleFunc("GET /v1/tasks", handlers.NewGetAllTasksHandler(logger, service))
	router.HandleFunc("POST /v1/tasks", handlers.NewPostTaskHandler(logger, service))
	router.HandleFunc("PUT /v1/tasks/{id}", handlers.NewTaskUpdater(logger, service))
	router.HandleFunc("DELETE /v1/tasks/{id}", handlers.NewDeleteTaskHandler(logger, service))

	return router
}
