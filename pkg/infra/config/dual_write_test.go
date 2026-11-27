package config

import (
	"os"
	"testing"
)

func TestDefaultDualWriteConfig(t *testing.T) {
	cfg := DefaultDualWriteConfig()
	if !cfg.Enabled {
		t.Fatal("expected new path enabled by default")
	}
	if cfg.FailurePolicy != DualWriteFailRequest {
		t.Fatalf("unexpected failure policy: %s", cfg.FailurePolicy)
	}
	if !cfg.IsNewPathOnly() {
		t.Fatal("expected Phase-4 new-path-only defaults")
	}
	ops := cfg.Operations()
	if len(ops) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(ops))
	}
}

func TestDualWriteConfigFromEnv(t *testing.T) {
	t.Setenv("MATCHMAKING_DUAL_WRITE_ENABLED", "false")
	t.Setenv("MATCHMAKING_DUAL_WRITE_FAILURE_POLICY", "legacy_only_and_alert")
	cfg := DualWriteConfigFromEnv()
	if cfg.Enabled {
		t.Fatal("expected new path disabled")
	}
	if cfg.FailurePolicy != DualWriteFailLegacyOnlyAndAlert {
		t.Fatalf("unexpected failure policy: %s", cfg.FailurePolicy)
	}

	_ = os.Unsetenv("MATCHMAKING_DUAL_WRITE_ENABLED")
	_ = os.Unsetenv("MATCHMAKING_DUAL_WRITE_FAILURE_POLICY")
	cfg = DualWriteConfigFromEnv()
	if !cfg.Enabled {
		t.Fatal("expected default enabled when env unset")
	}
	if cfg.FailurePolicy != DualWriteFailRequest {
		t.Fatalf("expected default fail_request, got %s", cfg.FailurePolicy)
	}
}
