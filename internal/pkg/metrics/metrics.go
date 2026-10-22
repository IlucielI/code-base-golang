package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry *prometheus.Registry
	once     sync.Once

	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge
)

func initMetrics() {
	registry = prometheus.NewRegistry()

	// Register standard Go runtime and process metrics collectors
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total count of processed HTTP requests partitioned by method, path, and status code.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency histogram of processed HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current count of active in-flight HTTP requests being processed.",
		},
	)

	registry.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsInFlight,
	)
}

// EnsureInit initializes collectors once.
func EnsureInit() {
	once.Do(initMetrics)
}

// Registry returns the active Prometheus registry.
func Registry() *prometheus.Registry {
	EnsureInit()
	return registry
}

// Handler returns the Prometheus HTTP metrics scraping handler.
func Handler() http.Handler {
	return promhttp.HandlerFor(Registry(), promhttp.HandlerOpts{})
}

// ObserveHTTPRequest records the HTTP request duration, increments total requests, and decrements in-flight.
func ObserveHTTPRequest(method, path, status string, durationSeconds float64) {
	EnsureInit()
	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	httpRequestDuration.WithLabelValues(method, path, status).Observe(durationSeconds)
}

// IncInFlight increments the in-flight requests gauge.
func IncInFlight() {
	EnsureInit()
	httpRequestsInFlight.Inc()
}

// DecInFlight decrements the in-flight requests gauge.
func DecInFlight() {
	EnsureInit()
	httpRequestsInFlight.Dec()
}
