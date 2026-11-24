package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/leet-gaming/match-making-api/pkg/infra/config"
)

// ConsistencyTicker runs ConsistencyChecker on a fixed interval (2509-002).
type ConsistencyTicker struct {
	checker *ConsistencyChecker
	config  config.ConsistencyCheckConfig
}

// NewConsistencyTicker creates a periodic consistency monitor.
func NewConsistencyTicker(checker *ConsistencyChecker, cfg config.ConsistencyCheckConfig) *ConsistencyTicker {
	if cfg.Interval <= 0 {
		cfg.Interval = 60 * time.Second
	}
	return &ConsistencyTicker{checker: checker, config: cfg}
}

// Start blocks until ctx is cancelled, running checks on each tick.
func (t *ConsistencyTicker) Start(ctx context.Context) error {
	if !t.config.Enabled {
		slog.Info("ConsistencyTicker is disabled, not starting")
		return nil
	}
	if t.checker == nil {
		slog.Error("ConsistencyTicker missing checker")
		return errString("consistency ticker missing checker")
	}

	slog.Info("Starting ConsistencyTicker",
		"interval", t.config.Interval,
		"drift_alert_threshold", t.config.DriftAlertThreshold)

	// Run once immediately so metrics appear without waiting a full interval.
	t.checker.RunOnce(ctx)

	ticker := time.NewTicker(t.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("ConsistencyTicker stopped (context cancelled)")
			return nil
		case <-ticker.C:
			t.checker.RunOnce(ctx)
		}
	}
}
