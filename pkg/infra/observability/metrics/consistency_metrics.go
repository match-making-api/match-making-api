package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ConsistencyChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchmaking_consistency_checks_total",
			Help: "Total consistency check runs by outcome (ok, drift, error).",
		},
		[]string{"outcome"},
	)

	ConsistencyDriftTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchmaking_consistency_drift_total",
			Help: "Total drift findings by kind (legacy_only_queue, new_only_queue, matched_still_queued).",
		},
		[]string{"kind"},
	)

	ConsistencyDriftGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "matchmaking_consistency_drift_current",
			Help: "Drift count observed in the latest successful check by kind.",
		},
		[]string{"kind"},
	)

	ConsistencyLastCheckUnix = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "matchmaking_consistency_last_check_timestamp_seconds",
			Help: "Unix timestamp of the last completed consistency check (success or error).",
		},
	)

	ConsistencyCheckDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "matchmaking_consistency_check_duration_seconds",
			Help:    "Duration of a consistency check run.",
			Buckets: prometheus.DefBuckets,
		},
	)
)
