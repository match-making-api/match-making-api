package config

import (
	"os"
	"testing"
)

func TestDefaultDualWriteConfig(t *testing.T) {
	cfg := DefaultDualWriteConfig()
	if !cfg.Enabled {
		t.Fatal("expected dual-write enabled by default")
	}
	if cfg.FailurePolicy != DualWriteFailLegacyOnlyAndAlert {
		t.Fatalf("unexpected failure policy: %s", cfg.FailurePolicy)
	}
	ops := cfg.Operations()
	if len(ops) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(ops))
	}
}

func TestDualWriteConfigFromEnv(t *testing.T) {
	t.Setenv("MATCHMAKING_DUAL_WRITE_ENABLED", "false")
	t.Setenv("MATCHMAKING_DUAL_WRITE_FAILURE_POLICY", "fail_request")
	cfg := DualWriteConfigFromEnv()
	if cfg.Enabled {
		t.Fatal("expected dual-write disabled")
	}
	if cfg.FailurePolicy != DualWriteFailRequest {
		t.Fatalf("unexpected failure policy: %s", cfg.FailurePolicy)
	}

	_ = os.Unsetenv("MATCHMAKING_DUAL_WRITE_ENABLED")
	_ = os.Unsetenv("MATCHMAKING_DUAL_WRITE_FAILURE_POLICY")
	cfg = DualWriteConfigFromEnv()
	if !cfg.Enabled {
		t.Fatal("expected default enabled when env unset")
	}
}
