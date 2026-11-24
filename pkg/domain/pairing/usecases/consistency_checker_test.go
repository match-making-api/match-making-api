package usecases

import (
	"context"
	"testing"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

type stubLegacyReader struct {
	sessions []pairing_out.LegacySessionRef
	err      error
}

func (s *stubLegacyReader) ListByStatuses(ctx context.Context, statuses []string) ([]pairing_out.LegacySessionRef, error) {
	if s.err != nil {
		return nil, s.err
	}
	allowed := make(map[string]struct{}, len(statuses))
	for _, st := range statuses {
		allowed[st] = struct{}{}
	}
	var out []pairing_out.LegacySessionRef
	for _, sess := range s.sessions {
		if _, ok := allowed[sess.Status]; ok {
			out = append(out, sess)
		}
	}
	return out, nil
}

func TestConsistencyChecker_NoDrift(t *testing.T) {
	pid := uuid.New()
	legacy := &stubLegacyReader{sessions: []pairing_out.LegacySessionRef{
		{PlayerID: pid, Status: LegacyStatusQueued},
	}}
	aq := pairing_entities.NewInMemoryActiveQueueStore()
	_ = aq.Register(context.Background(), &pairing_entities.ActiveQueueEntry{PlayerID: pid})

	checker := NewConsistencyChecker(legacy, aq, 1)
	report := checker.RunOnce(context.Background())
	if report.CheckError != nil {
		t.Fatalf("unexpected error: %v", report.CheckError)
	}
	if report.DriftCount != 0 {
		t.Fatalf("expected no drift, got %d", report.DriftCount)
	}
}

func TestConsistencyChecker_LegacyOnlyAndMatchedStillQueued(t *testing.T) {
	legacyOnly := uuid.New()
	matchedStill := uuid.New()
	aligned := uuid.New()

	legacy := &stubLegacyReader{sessions: []pairing_out.LegacySessionRef{
		{PlayerID: legacyOnly, Status: LegacyStatusSearching},
		{PlayerID: matchedStill, Status: LegacyStatusMatched},
		{PlayerID: aligned, Status: LegacyStatusQueued},
	}}
	aq := pairing_entities.NewInMemoryActiveQueueStore()
	_ = aq.Register(context.Background(), &pairing_entities.ActiveQueueEntry{PlayerID: matchedStill})
	_ = aq.Register(context.Background(), &pairing_entities.ActiveQueueEntry{PlayerID: aligned})

	checker := NewConsistencyChecker(legacy, aq, 1)
	report := checker.RunOnce(context.Background())
	if report.CheckError != nil {
		t.Fatalf("unexpected error: %v", report.CheckError)
	}
	if len(report.LegacyOnlyQueue) != 1 || report.LegacyOnlyQueue[0] != legacyOnly {
		t.Fatalf("legacy_only_queue = %v", report.LegacyOnlyQueue)
	}
	if len(report.MatchedStillQueued) != 1 || report.MatchedStillQueued[0] != matchedStill {
		t.Fatalf("matched_still_queued = %v", report.MatchedStillQueued)
	}
	if report.DriftCount != 2 {
		t.Fatalf("expected drift 2, got %d", report.DriftCount)
	}
}

func TestConsistencyChecker_NewOnlyQueue(t *testing.T) {
	newOnly := uuid.New()
	legacy := &stubLegacyReader{sessions: nil}
	aq := pairing_entities.NewInMemoryActiveQueueStore()
	_ = aq.Register(context.Background(), &pairing_entities.ActiveQueueEntry{PlayerID: newOnly})

	checker := NewConsistencyChecker(legacy, aq, 1)
	report := checker.RunOnce(context.Background())
	if len(report.NewOnlyQueue) != 1 || report.NewOnlyQueue[0] != newOnly {
		t.Fatalf("new_only_queue = %v", report.NewOnlyQueue)
	}
}

func TestConsistencyChecker_LegacyReadError(t *testing.T) {
	legacy := &stubLegacyReader{err: errString("mongo down")}
	aq := pairing_entities.NewInMemoryActiveQueueStore()
	checker := NewConsistencyChecker(legacy, aq, 1)
	report := checker.RunOnce(context.Background())
	if report.CheckError == nil {
		t.Fatal("expected check error")
	}
}
