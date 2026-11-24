package config

import (
	"os"
	"strconv"
	"strings"
)

// DualWriteFailurePolicy defines behaviour when the Kafka (new) path fails
// after the legacy / local write succeeded during Strimzi Migration Phase 2.
type DualWriteFailurePolicy string

const (
	// DualWriteFailLegacyOnlyAndAlert keeps the local/legacy write and logs/alerts
	// when Kafka produce (or new-path side effects) fail. Request is not failed.
	DualWriteFailLegacyOnlyAndAlert DualWriteFailurePolicy = "legacy_only_and_alert"

	// DualWriteFailRequest fails the operation when the new path fails (strict).
	DualWriteFailRequest DualWriteFailurePolicy = "fail_request"
)

// DualWriteConfig controls event-path participation during dual-write migration (2509-001).
//
// Producer-side dual-write (replay-api → legacy session + Kafka PlayerQueued/LeftQueue)
// is gated by KAFKA_MATCHMAKING_PRODUCER_ENABLED on replay-api.
// This config gates the match-making-api consumer / match-creation side of the new path.
type DualWriteConfig struct {
	// Enabled when true: process PlayerQueued / PlayerLeftQueue and publish MatchCreated
	// (new path active alongside legacy). When false: skip new-path handlers (rollback).
	Enabled bool

	// FailurePolicy applies when MatchCreated (or other dual-write publishes) fail
	// after local pool/pair state was already updated.
	FailurePolicy DualWriteFailurePolicy
}

// DefaultDualWriteConfig returns Phase-2 defaults: dual-write on, legacy-only-and-alert on produce failure.
func DefaultDualWriteConfig() DualWriteConfig {
	return DualWriteConfig{
		Enabled:       true,
		FailurePolicy: DualWriteFailLegacyOnlyAndAlert,
	}
}

// DualWriteConfigFromEnv loads dual-write settings from environment.
//
//	MATCHMAKING_DUAL_WRITE_ENABLED — default true
//	MATCHMAKING_DUAL_WRITE_FAILURE_POLICY — "legacy_only_and_alert" (default) | "fail_request"
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

// Operations lists matchmaking ops covered by dual-write during Phase 2.
func (c DualWriteConfig) Operations() []string {
	return []string{"queue_join", "queue_leave", "match_create"}
}
