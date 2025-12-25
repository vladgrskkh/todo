package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vladgrskkh/todo/internal/apierrors"
)

// RecoverPanic returns a middleware function that recovers from panics and
// returns a 500 Internal Server Error response to the client.
func RecoverPanic(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "Close")

					apierrors.ServerErrorResponse(logger, w, r, fmt.Errorf("%s", err))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
