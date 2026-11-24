package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TrafficNewPathPercent is the configured Phase-3 rollout percentage (0–100).
	TrafficNewPathPercent = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "matchmaking_traffic_new_path_percent",
			Help: "Configured percentage of traffic routed to the new Kafka / match-making-api path (2509-003).",
		},
	)

	// TrafficRouteDecisionsTotal counts routing decisions by path (new|legacy) for A/B quality comparison.
	TrafficRouteDecisionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchmaking_traffic_route_decisions_total",
			Help: "Traffic migration routing decisions by path (new, legacy).",
		},
		[]string{"path"},
	)
)
