package metrics

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultMetricsPort = 9090
)

var (
	MessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchmaking_kafka_messages_consumed_total",
			Help: "Total Kafka messages consumed by topic and consumer group.",
		},
		[]string{"topic", "group_id"},
	)

	MessagesProduced = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchmaking_kafka_messages_produced_total",
			Help: "Total Kafka messages produced by topic.",
		},
		[]string{"topic"},
	)

	ConsumerErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchmaking_kafka_consumer_errors_total",
			Help: "Total consumer processing or fetch errors by topic and group.",
		},
		[]string{"topic", "group_id", "error_type"},
	)

	ProducerErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchmaking_kafka_producer_errors_total",
			Help: "Total producer publish errors by topic.",
		},
		[]string{"topic"},
	)

	MessageProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "matchmaking_kafka_message_processing_seconds",
			Help:    "Time spent processing a consumed Kafka message.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic", "group_id"},
	)
)

// Handler exposes Prometheus metrics for scraping.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServerIfEnabled starts an HTTP server on METRICS_PORT (default 9090).
// Set METRICS_ENABLED=false to disable.
func StartServerIfEnabled() {
	if os.Getenv("METRICS_ENABLED") == "false" {
		return
	}

	port := defaultMetricsPort
	if v := os.Getenv("METRICS_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			port = p
		}
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", Handler())
		addr := ":" + strconv.Itoa(port)
		_ = http.ListenAndServe(addr, mux)
	}()
}

// ObserveProcessing records successful message processing duration.
func ObserveProcessing(topic, groupID string, start time.Time) {
	MessageProcessingDuration.WithLabelValues(topic, groupID).Observe(time.Since(start).Seconds())
}
