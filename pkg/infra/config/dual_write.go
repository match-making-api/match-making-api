package config

import (
	"os"
	"strconv"
	"strings"
)

// DualWriteFailurePolicy defines behaviour when MatchCreated (new path) fails
// after local pool/pair state was already updated.
type DualWriteFailurePolicy string

const (
	// DualWriteFailLegacyOnlyAndAlert keeps the local write and logs/alerts
	// when Kafka produce fails. Retained for emergency rollback only after Phase 4.
	DualWriteFailLegacyOnlyAndAlert DualWriteFailurePolicy = "legacy_only_and_alert"

	// DualWriteFailRequest fails the operation when the new path fails (strict).
	// Phase 4 (2509-004) default: new path is authoritative; dual-write soft-fail retired.
	DualWriteFailRequest DualWriteFailurePolicy = "fail_request"
)

// DualWriteConfig gates new-path consumer handlers after Strimzi migration.
//
// Phase 4 (2509-004) decommissioned dual-write as the operating mode: the Kafka /
// match-making-api path is the only active path. Enabled=false remains an emergency
// kill switch (see .docs/LEGACY_DECOMMISSION.md). Producer-side Kafka produce on
// replay-api is gated by KAFKA_MATCHMAKING_PRODUCER_ENABLED.
type DualWriteConfig struct {
	// Enabled when true: process PlayerQueued / PlayerLeftQueue and publish MatchCreated.
	// When false: skip new-path handlers (emergency rollback only after Phase 4).
	Enabled bool

	// FailurePolicy applies when MatchCreated publish fails after local pool/pair update.
	FailurePolicy DualWriteFailurePolicy
}

// DefaultDualWriteConfig returns Phase-4 defaults: new path on, fail_request on produce failure.
func DefaultDualWriteConfig() DualWriteConfig {
	return DualWriteConfig{
		Enabled:       true,
		FailurePolicy: DualWriteFailRequest,
	}
}

// DualWriteConfigFromEnv loads new-path settings from environment.
//
//	MATCHMAKING_DUAL_WRITE_ENABLED — default true (emergency false = new path off)
//	MATCHMAKING_DUAL_WRITE_FAILURE_POLICY — "fail_request" (default) | "legacy_only_and_alert"
func DualWriteConfigFromEnv() DualWriteConfig {
	cfg := DefaultDualWriteConfig()
	if v := os.Getenv("MATCHMAKING_DUAL_WRITE_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Enabled = b
		}
	}
	if v := strings.TrimSpace(os.Getenv("MATCHMAKING_DUAL_WRITE_FAILURE_POLICY")); v != "" {
		switch DualWriteFailurePolicy(strings.ToLower(v)) {
		case DualWriteFailLegacyOnlyAndAlert, DualWriteFailRequest:
			cfg.FailurePolicy = DualWriteFailurePolicy(strings.ToLower(v))
		}
	}
	return cfg
}

// Operations lists matchmaking ops on the new path (post-decommission sole path).
func (c DualWriteConfig) Operations() []string {
	return []string{"queue_join", "queue_leave", "match_create"}
}

// IsNewPathOnly reports Phase-4 operating mode: handlers on and strict produce failure.
func (c DualWriteConfig) IsNewPathOnly() bool {
	return c.Enabled && c.FailurePolicy == DualWriteFailRequest
}
