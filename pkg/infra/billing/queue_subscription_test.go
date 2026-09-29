package billing

import (
	"context"
	"errors"
	"testing"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/subscription"
	"google.golang.org/grpc"
)

type stubSubscriptionClient struct {
	resp *GetSubscriptionResponse
	err  error
}

func (s stubSubscriptionClient) GetSubscription(ctx context.Context, in *GetSubscriptionRequest, opts ...grpc.CallOption) (*GetSubscriptionResponse, error) {
	return s.resp, s.err
}

func (s stubSubscriptionClient) ValidateOperation(context.Context, *ValidateOperationRequest, ...grpc.CallOption) (*ValidateOperationResponse, error) {
	return nil, errors.New("ValidateOperation must not be called")
}

func (s stubSubscriptionClient) ConfirmOperation(context.Context, *ConfirmOperationRequest, ...grpc.CallOption) (*ConfirmOperationResponse, error) {
	return nil, errors.New("ConfirmOperation must not be called")
}

func TestQueueSubscriptionLookup_ReadOnly(t *testing.T) {
	lookup := NewQueueSubscriptionLookup(stubSubscriptionClient{
		resp: &GetSubscriptionResponse{
			IsValid: true,
			Subscription: &Subscription{
				PlanId:    "premium",
				Status:    "active",
				Available: map[string]int32{"queue_join": 2},
			},
		},
	})
	snap, err := lookup.Lookup(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if snap.PlanID != "premium" || snap.QueueJoinsRemaining == nil || *snap.QueueJoinsRemaining != 2 {
		t.Fatalf("snapshot: %+v", snap)
	}
}

func TestQueueSubscriptionLookup_Invalid(t *testing.T) {
	lookup := NewQueueSubscriptionLookup(stubSubscriptionClient{
		resp: &GetSubscriptionResponse{IsValid: false, Reason: "expired"},
	})
	_, err := lookup.Lookup(context.Background(), "user-1")
	if !errors.Is(err, subscription.ErrSubscriptionInactive) {
		t.Fatalf("got %v", err)
	}
}
