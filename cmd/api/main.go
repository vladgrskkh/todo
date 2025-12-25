package main

import (
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/vladgrskkh/todo/config"
	"github.com/vladgrskkh/todo/internal/handlers/routes"
	"github.com/vladgrskkh/todo/internal/repository"
	"github.com/vladgrskkh/todo/internal/server"
	"github.com/vladgrskkh/todo/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// parsing config
	cfg, err := config.New()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("creating task repository and todo service")
	taskRepo := repository.NewTaskRepo()
	service := service.NewTodoService(logger, taskRepo)

	logger.Info("creating routes and server")
	router := routes.Routes(logger, service)
	s := server.New(logger, cfg, router)

	logger.Info("starting server on port", slog.Int("port", cfg.Port))
	err = s.Serve()
	if err != nil {
		logger.Error(err.Error(), slog.String("trace", string(debug.Stack())))
		os.Exit(1)
	}
}

// TODO: mb make so that when task is done it cannot be updated
// TODO: repo persistance
// TODO: unit, integration tests
// TODO: load test
// TODO: metrics
// TODO: helthcheck
// TODO: panic recovery, logging middleware
// TODO: repo context
