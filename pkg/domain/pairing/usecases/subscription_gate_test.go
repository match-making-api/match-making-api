package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/subscription"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

type staticLookup struct {
	snap *subscription.Snapshot
	err  error
}

func (s staticLookup) Lookup(ctx context.Context, userID string) (*subscription.Snapshot, error) {
	return s.snap, s.err
}

func TestHandlePlayerQueuedProto_RejectsFreePriorityBeforePool(t *testing.T) {
	exec := &recordingAdd{}
	consumer := NewMatchmakingEventConsumer(exec, nil, nil, nil, nil, nil, nil)
	consumer.SetSubscriptionLookup(staticLookup{
		snap: &subscription.Snapshot{PlanID: "free", Status: "active"},
	})

	boost := int32(1)
	err := consumer.HandlePlayerQueuedProto(context.Background(),
		&schemas.EventEnvelope{ResourceOwnerId: "rid-1"},
		&schemas.PlayerQueuedPayload{
			PlayerId:      uuid.NewString(),
			GameId:        uuid.NewString(),
			Region:        "br",
			PriorityBoost: &boost,
		},
	)
	if !errors.Is(err, subscription.ErrUpgradeRequired) {
		t.Fatalf("got %v", err)
	}
	if exec.called {
		t.Fatal("pool insert must not run when subscription check fails")
	}
}

func TestHandlePlayerQueuedProto_RejectsUsageLimit(t *testing.T) {
	exec := &recordingAdd{}
	consumer := NewMatchmakingEventConsumer(exec, nil, nil, nil, nil, nil, nil)
	zero := 0
	consumer.SetSubscriptionLookup(staticLookup{
		snap: &subscription.Snapshot{PlanID: "premium", Status: "active", QueueJoinsRemaining: &zero},
	})

	err := consumer.HandlePlayerQueuedProto(context.Background(),
		&schemas.EventEnvelope{ResourceOwnerId: "rid-1"},
		&schemas.PlayerQueuedPayload{
			PlayerId: uuid.NewString(),
			GameId:   uuid.NewString(),
			Region:   "br",
		},
	)
	if !errors.Is(err, subscription.ErrUsageLimitExceeded) {
		t.Fatalf("got %v", err)
	}
	if exec.called {
		t.Fatal("pool insert must not run when usage limit is exceeded")
	}
}

type recordingAdd struct {
	called bool
}

func (r *recordingAdd) Execute(payload FindPairPayload) (*pairing_entities.Pair, *pairing_entities.Pool, int, error) {
	r.called = true
	return nil, nil, 0, nil
}
