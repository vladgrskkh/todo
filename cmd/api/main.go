package main

import (
	"flag"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/vladgrskkh/todo/config"
	"github.com/vladgrskkh/todo/internal/handlers/middleware/metrics"
	"github.com/vladgrskkh/todo/internal/handlers/routes"
	"github.com/vladgrskkh/todo/internal/repository"
	"github.com/vladgrskkh/todo/internal/server"
	"github.com/vladgrskkh/todo/internal/service"
	"github.com/vladgrskkh/todo/pkg/envload"
	"github.com/vladgrskkh/todo/pkg/inmemorydb"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var envPath string

	flag.StringVar(&envPath, "envpath", ".env", "set path to .env file")

	flag.Parse()

	logger.Info("loading environment variables")
	err := envload.Load(envPath, true)
	if err != nil {
		logger.Error(err.Error())
	}

	logger.Info("loading config")
	cfg, err := config.New()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := inmemorydb.Open(cfg.DBPath)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer func() {
		e := db.Close()
		if e != nil {
			logger.Error(e.Error())
		}
	}()

	logger.Info("database opened")
	logger.Info("creating task repository and todo service")
	taskRepo := repository.NewTaskRepo(db)
	service := service.NewTodoService(logger, taskRepo)

	logger.Info("creating routes and server")
	router := routes.Routes(logger, service, cfg.Env, cfg.Version)
	s := server.New(logger, cfg, router)

	logger.Info("initializing metrics")
	metrics.InitMetrics()

	logger.Info("starting server on port", slog.Int("port", cfg.Port))
	err = s.Serve()
	if err != nil {
		logger.Error(err.Error(), slog.String("trace", string(debug.Stack())))
		os.Exit(1)
	}
}
