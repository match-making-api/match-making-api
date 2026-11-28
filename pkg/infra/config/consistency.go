package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultConsistencyInterval      = 60 * time.Second
	defaultConsistencyDriftAlert    = 1
	defaultLegacySessionsCollection = "matchmaking_sessions"
)

// ConsistencyCheckConfig controls periodic legacy-vs-new consistency monitoring (2509-002).
//
// Phase 4 (2509-004) disables this by default: dual-write is retired and legacy session
// comparison is no longer part of the steady-state path. Re-enable only during emergency
// rollback windows (see .docs/LEGACY_DECOMMISSION.md).
type ConsistencyCheckConfig struct {
	Enabled             bool
	Interval            time.Duration
	DriftAlertThreshold int
	LegacyCollection    string
	LegacyDBName        string // empty => use Config.MongoDB.DBName
}

// DefaultConsistencyCheckConfig returns Phase-4 defaults: consistency check off.
func DefaultConsistencyCheckConfig() ConsistencyCheckConfig {
	return ConsistencyCheckConfig{
		Enabled:             false,
		Interval:            defaultConsistencyInterval,
		DriftAlertThreshold: defaultConsistencyDriftAlert,
		LegacyCollection:    defaultLegacySessionsCollection,
	}
}

// ConsistencyCheckConfigFromEnv loads consistency-check settings from environment.
//
//	MATCHMAKING_CONSISTENCY_CHECK_ENABLED — default false (Phase 4)
//	MATCHMAKING_CONSISTENCY_CHECK_INTERVAL — Go duration, default 60s
//	MATCHMAKING_CONSISTENCY_DRIFT_ALERT_THRESHOLD — default 1
//	MATCHMAKING_LEGACY_SESSIONS_COLLECTION — default matchmaking_sessions
//	MATCHMAKING_LEGACY_SESSIONS_DB — optional override of MONGO_DB_NAME
func ConsistencyCheckConfigFromEnv() ConsistencyCheckConfig {
	cfg := DefaultConsistencyCheckConfig()
	if v := os.Getenv("MATCHMAKING_CONSISTENCY_CHECK_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Enabled = b
		}
	}
	if v := strings.TrimSpace(os.Getenv("MATCHMAKING_CONSISTENCY_CHECK_INTERVAL")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.Interval = d
		}
	}
	if v := strings.TrimSpace(os.Getenv("MATCHMAKING_CONSISTENCY_DRIFT_ALERT_THRESHOLD")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.DriftAlertThreshold = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("MATCHMAKING_LEGACY_SESSIONS_COLLECTION")); v != "" {
		cfg.LegacyCollection = v
	}
	if v := strings.TrimSpace(os.Getenv("MATCHMAKING_LEGACY_SESSIONS_DB")); v != "" {
		cfg.LegacyDBName = v
	}
	return cfg
}
