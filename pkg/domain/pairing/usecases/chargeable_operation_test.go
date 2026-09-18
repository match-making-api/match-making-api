package usecases

import (
	"os"
	"testing"
)

func TestPriorityBoostAmountCents_Default(t *testing.T) {
	_ = os.Unsetenv("PRIORITY_BOOST_AMOUNT_CENTS")
	if got := PriorityBoostAmountCents(1); got != 100 {
		t.Fatalf("want 100, got %d", got)
	}
}

func TestPriorityBoostAmountCents_Env(t *testing.T) {
	t.Setenv("PRIORITY_BOOST_AMOUNT_CENTS", "250")
	if got := PriorityBoostAmountCents(1); got != 250 {
		t.Fatalf("want 250, got %d", got)
	}
}

func TestPriorityBoostAmountCents_BoostLevelAsCents(t *testing.T) {
	if got := PriorityBoostAmountCents(500); got != 500 {
		t.Fatalf("want 500, got %d", got)
	}
}

func TestBuildQueuePriorityBoostEvent(t *testing.T) {
	ev := BuildQueuePriorityBoostEvent(
		"p1", "g1", "br", "t1", "c1", "rid", "corr-1", 1,
	)
	if ev.Type != "ChargeableOperationRequested" {
		t.Fatalf("type: %s", ev.Type)
	}
	p := ev.ChargeableOperationRequested
	if p == nil || p.OperationType != "queue_priority_boost" {
		t.Fatalf("payload: %+v", p)
	}
	if p.AmountCents != 100 || p.PlayerID != "p1" || p.CorrelationID != "corr-1" {
		t.Fatalf("fields: %+v", p)
	}
	if p.IdempotencyKey == "" {
		t.Fatal("idempotency_key required")
	}
}
