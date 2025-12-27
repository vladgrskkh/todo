package metrics

import (
	"expvar"
	"net/http"
	"strconv"
	"time"
)

var (
	totalRequests     *expvar.Int
	totalResponses    *expvar.Int
	totalLatencyMs    *expvar.Int
	statusCounts      *expvar.Map
	TotalTasksCreated *expvar.Int
	TotalTasksDone    *expvar.Int
)

// Wrapped for http.ResponseWriter.
// Main purpose is to track status code. In this application there will be no problems.
// But when intoducing more complex logic(with external dependencies/using different interfaces),
// it can cause bugs.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func InitMetrics() {
	totalRequests = expvar.NewInt("total_requests")
	totalResponses = expvar.NewInt("total_responses")
	totalLatencyMs = expvar.NewInt("total_latency_ms")
	statusCounts = expvar.NewMap("status_counts")

	// business metrics
	TotalTasksCreated = expvar.NewInt("total_tasks_created")
	TotalTasksDone = expvar.NewInt("total_tasks_done")
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequests.Add(1)

		rw := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		statusCounts.Add(strconv.Itoa(rw.statusCode), 1)
		totalResponses.Add(1)
		totalLatencyMs.Add(time.Since(start).Milliseconds())
	})
}
