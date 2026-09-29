package subscription

import (
	"errors"
	"strings"
)

// Tier is the queue-routing tier derived from a subscription plan.
type Tier string

const (
	TierFree    Tier = "free"
	TierPremium Tier = "premium"
)

// Snapshot is a read-only view of a player's subscription.
// Match-making does not update usage; wallet-api owns that.
type Snapshot struct {
	PlanID string
	Status string
	// QueueJoinsRemaining is set when the subscription API reports a limit.
	// Nil means no limit is defined for queue joins.
	QueueJoinsRemaining *int
}

var (
	// ErrUpgradeRequired is returned when a free tier requests a premium-only option.
	ErrUpgradeRequired = errors.New("upgrade required")
	// ErrUsageLimitExceeded is returned when queue join usage has no remaining allowance.
	ErrUsageLimitExceeded = errors.New("usage limit exceeded")
	// ErrSubscriptionInactive is returned when the subscription is missing or not active.
	ErrSubscriptionInactive = errors.New("subscription inactive")
)

// TierOf maps a plan id to free or premium.
// Premium covers plan ids that contain premium, priority, plus, or equal pro.
func TierOf(planID string) Tier {
	p := strings.ToLower(strings.TrimSpace(planID))
	switch {
	case p == "pro", strings.Contains(p, "premium"), strings.Contains(p, "priority"), strings.Contains(p, "plus"):
		return TierPremium
	default:
		return TierFree
	}
}

// CheckQueueJoin validates tier and usage before a player is added to the pool.
// A nil snapshot allows the basic queue and rejects priority boost.
func CheckQueueJoin(snap *Snapshot, priorityBoost bool) error {
	if snap == nil {
		if priorityBoost {
			return ErrUpgradeRequired
		}
		return nil
	}

	status := strings.ToLower(strings.TrimSpace(snap.Status))
	if status != "" && status != "active" && status != "valid" {
		return ErrSubscriptionInactive
	}

	if snap.QueueJoinsRemaining != nil && *snap.QueueJoinsRemaining <= 0 {
		return ErrUsageLimitExceeded
	}

	if priorityBoost && TierOf(snap.PlanID) != TierPremium {
		return ErrUpgradeRequired
	}
	return nil
}
