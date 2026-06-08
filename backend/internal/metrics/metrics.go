package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	WSConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ws_connections_active",
		Help: "Number of active WebSocket connections",
	})

	MessagesSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_sent_total",
		Help: "Total messages sent via WebSocket",
	})
)
