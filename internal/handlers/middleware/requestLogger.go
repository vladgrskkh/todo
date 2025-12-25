package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// RequestLogger returns a middleware function that logs the request
// method, path, remote address and duration after the request is completed.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			logger.Info("request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("duration", time.Since(start).String()))
		})
	}
}
