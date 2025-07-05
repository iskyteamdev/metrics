package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func setupPrometheus() {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg
}

func TestInitRegistersMetrics(t *testing.T) {
	setupPrometheus()
	Init()

	if err := testutil.CollectAndCompare(RequestCounter, strings.NewReader(""), "http_requests_total"); err != nil {
		t.Errorf("RequestCounter not registered correctly: %v", err)
	}

	if err := testutil.CollectAndCompare(RequestDuration, strings.NewReader(""), "http_request_duration_seconds"); err != nil {
		t.Errorf("RequestDuration not registered correctly: %v", err)
	}
}

func TestMiddlewareAndHandler(t *testing.T) {
	setupPrometheus()
	Init()

	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, w.Code)
	}

	mh := Handler()
	req2 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w2 := httptest.NewRecorder()
	mh.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("metrics handler status = %d, want %d", w2.Code, http.StatusOK)
	}
	if !strings.Contains(w2.Body.String(), "http_requests_total") {
		t.Errorf("metrics output missing http_requests_total; got:\n%s", w2.Body.String())
	}
}
