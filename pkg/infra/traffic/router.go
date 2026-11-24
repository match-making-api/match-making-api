package traffic

import (
	"hash/fnv"
	"strings"

	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"github.com/leet-gaming/match-making-api/pkg/infra/observability/metrics"
)

const (
	PathNew    = "new"
	PathLegacy = "legacy"
)

// Decision is the routing outcome for one request.
type Decision struct {
	Path   string
	Reason string
}

// Router assigns traffic to new vs legacy paths for Phase 3 gradual rollout (2509-003).
type Router struct {
	cfg config.TrafficMigrationConfig
}

// NewRouter builds a router from traffic migration config.
func NewRouter(cfg config.TrafficMigrationConfig) *Router {
	return &Router{cfg: cfg}
}

// Decide returns whether the request should use the new Kafka / match-making-api path.
// key should be a sticky identifier (player_id). region is used for region strategy.
func (r *Router) Decide(key, region string) Decision {
	d := r.decide(key, region)
	metrics.TrafficRouteDecisionsTotal.WithLabelValues(d.Path).Inc()
	return d
}

// ShouldUseNewPath is a convenience wrapper around Decide.
func (r *Router) ShouldUseNewPath(key, region string) bool {
	return r.Decide(key, region).Path == PathNew
}

func (r *Router) decide(key, region string) Decision {
	if r.cfg.NewPathPercent <= 0 {
		return Decision{Path: PathLegacy, Reason: "percent_zero"}
	}

	switch r.cfg.Strategy {
	case config.TrafficStrategyRegion:
		if inAllowlist(region, r.cfg.RegionAllowlist) {
			return Decision{Path: PathNew, Reason: "region_allowlist"}
		}
		return Decision{Path: PathLegacy, Reason: "region_not_allowlisted"}
	default:
		if r.cfg.NewPathPercent >= 100 {
			return Decision{Path: PathNew, Reason: "percent_full"}
		}
		if stickyInPercent(key, r.cfg.NewPathPercent) {
			return Decision{Path: PathNew, Reason: "percent_bucket"}
		}
		return Decision{Path: PathLegacy, Reason: "percent_bucket"}
	}
}

// stickyInPercent returns true when fnv32(key)%100 is in [0, percent).
func stickyInPercent(key string, percent int) bool {
	if percent <= 0 {
		return false
	}
	if percent >= 100 {
		return true
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32()%100) < percent
}

func inAllowlist(region string, allowlist []string) bool {
	region = strings.ToLower(strings.TrimSpace(region))
	if region == "" {
		return false
	}
	for _, a := range allowlist {
		if region == a {
			return true
		}
	}
	return false
}

// SyncConfiguredPercentGauge publishes the configured rollout percent for dashboards.
func SyncConfiguredPercentGauge(cfg config.TrafficMigrationConfig) {
	metrics.TrafficNewPathPercent.Set(float64(cfg.NewPathPercent))
}
