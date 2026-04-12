package httpapi

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricSegmentInt = regexp.MustCompile(`^\d+$`)
	metricSegmentKey = regexp.MustCompile(`^[a-z0-9][a-z0-9\-._]{6,}$`)
)

type Metrics struct {
	registry         *prometheus.Registry
	inFlightRequests prometheus.Gauge
	httpRequests     *prometheus.CounterVec
	httpDuration     *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	inFlight := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "openclaw",
		Subsystem: "http",
		Name:      "in_flight_requests",
		Help:      "Current in-flight HTTP requests.",
	})
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "openclaw",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total HTTP requests handled by the platform API.",
	}, []string{"method", "route", "status"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "openclaw",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "route", "status"})

	registry.MustRegister(inFlight, requests, duration)
	return &Metrics{
		registry:         registry,
		inFlightRequests: inFlight,
		httpRequests:     requests,
		httpDuration:     duration,
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) Observe(req *http.Request, status int, duration time.Duration) {
	if m == nil {
		return
	}
	route := normalizeMetricRoute(req.URL.Path)
	statusLabel := http.StatusText(status)
	if statusLabel == "" {
		statusLabel = "unknown"
	}
	method := req.Method
	m.httpRequests.WithLabelValues(method, route, statusLabel).Inc()
	m.httpDuration.WithLabelValues(method, route, statusLabel).Observe(duration.Seconds())
}

func normalizeMetricRoute(path string) string {
	if path == "" {
		return "/"
	}
	parts := strings.Split(path, "/")
	for index, part := range parts {
		if part == "" {
			continue
		}
		switch {
		case metricSegmentInt.MatchString(part):
			parts[index] = ":id"
		case metricSegmentKey.MatchString(part) && !strings.Contains(part, "."):
			parts[index] = ":key"
		}
	}
	return strings.Join(parts, "/")
}

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newStatusCapturingResponseWriter(w http.ResponseWriter) *statusCapturingResponseWriter {
	return &statusCapturingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (w *statusCapturingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusCapturingResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
