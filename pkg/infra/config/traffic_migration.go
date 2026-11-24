package config

import (
	"os"
	"strconv"
	"strings"
)

// TrafficStrategy selects how requests are assigned to the new Kafka path vs legacy.
type TrafficStrategy string

const (
	// TrafficStrategyPercent sticky-hashes a routing key (usually player_id) into 0–99 buckets.
	TrafficStrategyPercent TrafficStrategy = "percent"
	// TrafficStrategyRegion sends allowlisted regions to the new path; others stay legacy.
	TrafficStrategyRegion TrafficStrategy = "region"
)

// TrafficMigrationConfig controls Phase 3 gradual rollout / A/B traffic split (2509-003).
//
// Producer-side enforcement lives on replay-api
// (KAFKA_MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT and related). This config mirrors the
// same knobs so match-making-api can log expected rollout and expose metrics.
type TrafficMigrationConfig struct {
	// NewPathPercent is 0–100. 0 = full rollback (legacy only); 100 = all eligible traffic on new path.
	NewPathPercent int

	// Strategy is "percent" (default) or "region".
	Strategy TrafficStrategy

	// RegionAllowlist is used when Strategy == region (comma-separated env → slice).
	RegionAllowlist []string
}

// DefaultTrafficMigrationConfig returns Phase-3 defaults: 100% new path, percent strategy.
// Ops lower NewPathPercent for gradual rollout (e.g. 10 → 50 → 100).
func DefaultTrafficMigrationConfig() TrafficMigrationConfig {
	return TrafficMigrationConfig{
		NewPathPercent:  100,
		Strategy:        TrafficStrategyPercent,
		RegionAllowlist: nil,
	}
}

// TrafficMigrationConfigFromEnv loads traffic migration settings from environment.
//
//	MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT — 0–100, default 100
//	MATCHMAKING_TRAFFIC_STRATEGY — "percent" (default) | "region"
//	MATCHMAKING_TRAFFIC_REGION_ALLOWLIST — comma-separated regions (region strategy)
func TrafficMigrationConfigFromEnv() TrafficMigrationConfig {
	cfg := DefaultTrafficMigrationConfig()
	if v := strings.TrimSpace(os.Getenv("MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.NewPathPercent = clampPercent(n)
		}
	}
	if v := strings.TrimSpace(os.Getenv("MATCHMAKING_TRAFFIC_STRATEGY")); v != "" {
		switch TrafficStrategy(strings.ToLower(v)) {
		case TrafficStrategyPercent, TrafficStrategyRegion:
			cfg.Strategy = TrafficStrategy(strings.ToLower(v))
		}
	}
	if v := strings.TrimSpace(os.Getenv("MATCHMAKING_TRAFFIC_REGION_ALLOWLIST")); v != "" {
		parts := strings.Split(v, ",")
		cfg.RegionAllowlist = make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(strings.ToLower(p))
			if p != "" {
				cfg.RegionAllowlist = append(cfg.RegionAllowlist, p)
			}
		}
	}
	return cfg
}

func clampPercent(n int) int {
	if n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
}

// IsFullRollback reports whether new-path traffic is fully disabled (0%).
func (c TrafficMigrationConfig) IsFullRollback() bool {
	return c.NewPathPercent <= 0
}

// IsFullRollout reports whether all eligible traffic targets the new path (100%).
func (c TrafficMigrationConfig) IsFullRollout() bool {
	return c.NewPathPercent >= 100 && c.Strategy == TrafficStrategyPercent
}
