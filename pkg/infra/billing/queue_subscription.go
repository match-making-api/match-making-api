package billing

import (
	"context"
	"fmt"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/subscription"
)

// QueueSubscriptionLookup reads GetSubscription only.
// It does not call ValidateOperation or ConfirmOperation and does not update usage.
type QueueSubscriptionLookup struct {
	client SubscriptionServiceClient
}

// NewQueueSubscriptionLookup builds a read-only lookup. A nil client skips the RPC.
func NewQueueSubscriptionLookup(client SubscriptionServiceClient) *QueueSubscriptionLookup {
	return &QueueSubscriptionLookup{client: client}
}

// Lookup returns the subscription snapshot for queue-join gating.
// Invalid or missing subscriptions return ErrSubscriptionInactive.
func (l *QueueSubscriptionLookup) Lookup(ctx context.Context, userID string) (*subscription.Snapshot, error) {
	if l == nil || l.client == nil {
		return nil, nil
	}
	resp, err := l.client.GetSubscription(ctx, &GetSubscriptionRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	if resp == nil || !resp.GetIsValid() || resp.GetSubscription() == nil {
		reason := "subscription invalid"
		if resp != nil && resp.GetReason() != "" {
			reason = resp.GetReason()
		}
		return nil, fmt.Errorf("%w: %s", subscription.ErrSubscriptionInactive, reason)
	}
	sub := resp.GetSubscription()
	snap := &subscription.Snapshot{
		PlanID: sub.GetPlanId(),
		Status: sub.GetStatus(),
	}
	if sub.Available != nil {
		if v, ok := sub.Available["queue_join"]; ok {
			n := int(v)
			snap.QueueJoinsRemaining = &n
		}
	}
	return snap, nil
}
