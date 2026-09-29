package subscription

import (
	"errors"
	"testing"
)

func TestCheckQueueJoin_BasicFreeAllowed(t *testing.T) {
	snap := &Snapshot{PlanID: "free", Status: "active"}
	if err := CheckQueueJoin(snap, false); err != nil {
		t.Fatal(err)
	}
}

func TestCheckQueueJoin_PriorityRequiresPremium(t *testing.T) {
	err := CheckQueueJoin(&Snapshot{PlanID: "free", Status: "active"}, true)
	if !errors.Is(err, ErrUpgradeRequired) {
		t.Fatalf("got %v", err)
	}
	if err := CheckQueueJoin(&Snapshot{PlanID: "premium-monthly", Status: "active"}, true); err != nil {
		t.Fatal(err)
	}
}

func TestCheckQueueJoin_NilSnapshotGatesPriorityOnly(t *testing.T) {
	if err := CheckQueueJoin(nil, false); err != nil {
		t.Fatal(err)
	}
	if err := CheckQueueJoin(nil, true); !errors.Is(err, ErrUpgradeRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestCheckQueueJoin_UsageLimit(t *testing.T) {
	zero := 0
	err := CheckQueueJoin(&Snapshot{PlanID: "premium", Status: "active", QueueJoinsRemaining: &zero}, false)
	if !errors.Is(err, ErrUsageLimitExceeded) {
		t.Fatalf("got %v", err)
	}
	one := 1
	if err := CheckQueueJoin(&Snapshot{PlanID: "free", Status: "active", QueueJoinsRemaining: &one}, false); err != nil {
		t.Fatal(err)
	}
}

func TestCheckQueueJoin_Inactive(t *testing.T) {
	err := CheckQueueJoin(&Snapshot{PlanID: "premium", Status: "cancelled"}, false)
	if !errors.Is(err, ErrSubscriptionInactive) {
		t.Fatalf("got %v", err)
	}
}

func TestTierOf(t *testing.T) {
	if TierOf("Pro") != TierPremium || TierOf("basic") != TierFree {
		t.Fatal("tier mapping")
	}
}
