package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Redirect metrics
	RedirectRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redirect_requests_total",
			Help: "Total redirect requests by status",
		},
		[]string{"status"},
	)

	RedirectLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redirect_latency_seconds",
			Help:    "Redirect latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1},
		},
		[]string{},
	)

	// Cache metrics
	CacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total cache hits",
		},
	)

	CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total cache misses",
		},
	)

	// Database metrics
	DBQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total database queries by operation",
		},
		[]string{"operation"},
	)

	DBLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_latency_seconds",
			Help:    "Database query latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
		},
		[]string{"operation"},
	)

	// Worker metrics
	ClickEventsEnqueued = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "click_events_enqueued_total",
			Help: "Total click events enqueued",
		},
	)

	ClickEventsProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "click_events_processed_total",
			Help: "Total click events processed",
		},
	)

	StreamLag = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "stream_lag_messages",
			Help: "Number of pending messages in stream",
		},
	)
)
