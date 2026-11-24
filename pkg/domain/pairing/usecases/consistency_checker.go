package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/observability/metrics"
)

const (
	LegacyStatusQueued    = "queued"
	LegacyStatusSearching = "searching"
	LegacyStatusMatched   = "matched"

	DriftKindLegacyOnlyQueue    = "legacy_only_queue"
	DriftKindNewOnlyQueue       = "new_only_queue"
	DriftKindMatchedStillQueued = "matched_still_queued"
)

// ConsistencyReport summarizes one consistency check between legacy sessions and the new-path active queue.
type ConsistencyReport struct {
	CheckedAt          time.Time
	LegacyQueuedCount  int
	NewQueuedCount     int
	LegacyMatchedCount int
	LegacyOnlyQueue    []uuid.UUID
	NewOnlyQueue       []uuid.UUID
	MatchedStillQueued []uuid.UUID
	DriftCount         int
	CheckError         error
}

// ConsistencyChecker compares legacy Mongo sessions with ActiveQueueStore (new path).
type ConsistencyChecker struct {
	legacyReader pairing_out.LegacySessionReader
	activeQueue  pairing_out.ActiveQueueStore
	alertThresh  int
}

// NewConsistencyChecker builds a checker. alertThresh is the drift count that triggers alert logs (default 1).
func NewConsistencyChecker(
	legacyReader pairing_out.LegacySessionReader,
	activeQueue pairing_out.ActiveQueueStore,
	alertThresh int,
) *ConsistencyChecker {
	if alertThresh < 0 {
		alertThresh = 1
	}
	return &ConsistencyChecker{
		legacyReader: legacyReader,
		activeQueue:  activeQueue,
		alertThresh:  alertThresh,
	}
}

// RunOnce performs one comparison cycle and updates metrics / alert logs.
func (c *ConsistencyChecker) RunOnce(ctx context.Context) ConsistencyReport {
	start := time.Now()
	report := ConsistencyReport{CheckedAt: start}
	defer func() {
		metrics.ConsistencyLastCheckUnix.Set(float64(time.Now().Unix()))
		metrics.ConsistencyCheckDuration.Observe(time.Since(start).Seconds())
	}()

	if c.legacyReader == nil || c.activeQueue == nil {
		report.CheckError = errConsistencyDeps
		metrics.ConsistencyChecksTotal.WithLabelValues("error").Inc()
		slog.ErrorContext(ctx, "consistency check failed: missing dependencies")
		return report
	}

	legacySessions, err := c.legacyReader.ListByStatuses(ctx, []string{
		LegacyStatusQueued, LegacyStatusSearching, LegacyStatusMatched,
	})
	if err != nil {
		report.CheckError = err
		metrics.ConsistencyChecksTotal.WithLabelValues("error").Inc()
		slog.ErrorContext(ctx, "consistency check failed reading legacy sessions",
			"error", err,
			"alert", "MatchmakingConsistencyCheckFailed")
		return report
	}

	newEntries, err := c.activeQueue.GetAll(ctx)
	if err != nil {
		report.CheckError = err
		metrics.ConsistencyChecksTotal.WithLabelValues("error").Inc()
		slog.ErrorContext(ctx, "consistency check failed reading active queue",
			"error", err,
			"alert", "MatchmakingConsistencyCheckFailed")
		return report
	}

	legacyQueued := make(map[uuid.UUID]struct{})
	legacyMatched := make(map[uuid.UUID]struct{})
	for _, s := range legacySessions {
		switch s.Status {
		case LegacyStatusQueued, LegacyStatusSearching:
			legacyQueued[s.PlayerID] = struct{}{}
		case LegacyStatusMatched:
			legacyMatched[s.PlayerID] = struct{}{}
		}
	}
	report.LegacyQueuedCount = len(legacyQueued)
	report.LegacyMatchedCount = len(legacyMatched)

	newQueued := make(map[uuid.UUID]struct{}, len(newEntries))
	for _, e := range newEntries {
		if e == nil {
			continue
		}
		newQueued[e.PlayerID] = struct{}{}
	}
	report.NewQueuedCount = len(newQueued)

	for id := range legacyQueued {
		if _, ok := newQueued[id]; !ok {
			report.LegacyOnlyQueue = append(report.LegacyOnlyQueue, id)
		}
	}
	for id := range newQueued {
		if _, ok := legacyQueued[id]; !ok {
			// Matched-in-legacy but still in new queue is reported separately.
			if _, matched := legacyMatched[id]; matched {
				continue
			}
			report.NewOnlyQueue = append(report.NewOnlyQueue, id)
		}
	}
	for id := range legacyMatched {
		if _, ok := newQueued[id]; ok {
			report.MatchedStillQueued = append(report.MatchedStillQueued, id)
		}
	}

	report.DriftCount = len(report.LegacyOnlyQueue) + len(report.NewOnlyQueue) + len(report.MatchedStillQueued)

	metrics.ConsistencyDriftGauge.WithLabelValues(DriftKindLegacyOnlyQueue).Set(float64(len(report.LegacyOnlyQueue)))
	metrics.ConsistencyDriftGauge.WithLabelValues(DriftKindNewOnlyQueue).Set(float64(len(report.NewOnlyQueue)))
	metrics.ConsistencyDriftGauge.WithLabelValues(DriftKindMatchedStillQueued).Set(float64(len(report.MatchedStillQueued)))

	if len(report.LegacyOnlyQueue) > 0 {
		metrics.ConsistencyDriftTotal.WithLabelValues(DriftKindLegacyOnlyQueue).Add(float64(len(report.LegacyOnlyQueue)))
	}
	if len(report.NewOnlyQueue) > 0 {
		metrics.ConsistencyDriftTotal.WithLabelValues(DriftKindNewOnlyQueue).Add(float64(len(report.NewOnlyQueue)))
	}
	if len(report.MatchedStillQueued) > 0 {
		metrics.ConsistencyDriftTotal.WithLabelValues(DriftKindMatchedStillQueued).Add(float64(len(report.MatchedStillQueued)))
	}

	if report.DriftCount == 0 {
		metrics.ConsistencyChecksTotal.WithLabelValues("ok").Inc()
		slog.InfoContext(ctx, "consistency check ok",
			"legacy_queued", report.LegacyQueuedCount,
			"new_queued", report.NewQueuedCount,
			"legacy_matched", report.LegacyMatchedCount)
		return report
	}

	metrics.ConsistencyChecksTotal.WithLabelValues("drift").Inc()
	slog.WarnContext(ctx, "consistency drift detected",
		"drift_count", report.DriftCount,
		"legacy_only_queue", len(report.LegacyOnlyQueue),
		"new_only_queue", len(report.NewOnlyQueue),
		"matched_still_queued", len(report.MatchedStillQueued),
		"legacy_queued", report.LegacyQueuedCount,
		"new_queued", report.NewQueuedCount,
		"legacy_matched", report.LegacyMatchedCount,
		"alert", "MatchmakingConsistencyDrift")

	if report.DriftCount >= c.alertThresh {
		slog.ErrorContext(ctx, "consistency drift above alert threshold",
			"drift_count", report.DriftCount,
			"threshold", c.alertThresh,
			"legacy_only_sample", sampleIDs(report.LegacyOnlyQueue, 5),
			"new_only_sample", sampleIDs(report.NewOnlyQueue, 5),
			"matched_still_queued_sample", sampleIDs(report.MatchedStillQueued, 5),
			"alert", "MatchmakingConsistencyDrift",
			"runbook", "consistency-alert")
	}

	return report
}

var errConsistencyDeps = errString("consistency checker missing legacy reader or active queue store")

type errString string

func (e errString) Error() string { return string(e) }

func sampleIDs(ids []uuid.UUID, n int) []string {
	if n > len(ids) {
		n = len(ids)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, ids[i].String())
	}
	return out
}
