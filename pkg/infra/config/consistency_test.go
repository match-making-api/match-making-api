package config

import (
	"testing"
	"time"
)

func TestDefaultConsistencyCheckConfig(t *testing.T) {
	cfg := DefaultConsistencyCheckConfig()
	if cfg.Enabled {
		t.Fatal("expected consistency check disabled by default after Phase 4")
	}
	if cfg.Interval != 60*time.Second {
		t.Fatalf("expected 60s interval, got %v", cfg.Interval)
	}
	if cfg.DriftAlertThreshold != 1 {
		t.Fatalf("expected threshold 1, got %d", cfg.DriftAlertThreshold)
	}
	if cfg.LegacyCollection != "matchmaking_sessions" {
		t.Fatalf("unexpected collection %q", cfg.LegacyCollection)
	}
}

func TestConsistencyCheckConfigFromEnv(t *testing.T) {
	t.Setenv("MATCHMAKING_CONSISTENCY_CHECK_ENABLED", "true")
	t.Setenv("MATCHMAKING_CONSISTENCY_CHECK_INTERVAL", "30s")
	t.Setenv("MATCHMAKING_CONSISTENCY_DRIFT_ALERT_THRESHOLD", "5")
	t.Setenv("MATCHMAKING_LEGACY_SESSIONS_COLLECTION", "sessions_v2")
	t.Setenv("MATCHMAKING_LEGACY_SESSIONS_DB", "replay")

	cfg := ConsistencyCheckConfigFromEnv()
	if !cfg.Enabled {
		t.Fatal("expected enabled when env true")
	}
	if cfg.Interval != 30*time.Second {
		t.Fatalf("expected 30s, got %v", cfg.Interval)
	}
	if cfg.DriftAlertThreshold != 5 {
		t.Fatalf("expected threshold 5, got %d", cfg.DriftAlertThreshold)
	}
	if cfg.LegacyCollection != "sessions_v2" {
		t.Fatalf("unexpected collection %q", cfg.LegacyCollection)
	}
	if cfg.LegacyDBName != "replay" {
		t.Fatalf("unexpected db %q", cfg.LegacyDBName)
	}
}
