package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Counter of HTTP requests by method, path and status code
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Histogram of HTTP request durations (seconds)
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request latencies",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// Init registers all metrics in the default Prometheus registry.
func Init() {
	prometheus.MustRegister(RequestCounter, RequestDuration)
}

// Middleware is an HTTP middleware that records metrics.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		RequestDuration.
			WithLabelValues(r.Method, r.URL.Path).
			Observe(duration)

		RequestCounter.
			WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.statusCode)).
			Inc()
	})
}

// Handler returns the HTTP handler for Prometheus to scrape.
func Handler() http.Handler {
	return promhttp.Handler()
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
