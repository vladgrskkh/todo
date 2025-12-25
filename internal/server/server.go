package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vladgrskkh/todo/config"
)

type Server struct {
	logger *slog.Logger
	srv    *http.Server
}

func New(logger *slog.Logger, cfg *config.Config, routes http.Handler) *Server {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      routes,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return &Server{
		logger: nil,
		srv:    srv,
	}
}

func (s *Server) Serve() error {
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		// catch signals
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		sig := <-quit

		s.logger.Info("shutting down server", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := s.srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		shutdownError <- nil

	}()

	err := s.srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error while starting server: %w", err)
	}

	err = <-shutdownError
	if err != nil {
		return fmt.Errorf("error while shutting down server: %w", err)
	}

	s.logger.Info("server stopped", slog.String("addr", s.srv.Addr))
	return nil
}
