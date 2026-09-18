package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// ChargeablePublisher publishes ChargeableOperationRequested (wallet consumes).
type ChargeablePublisher interface {
	PublishChargeableOperationRequested(ctx context.Context, event *schemas.ChargeableOperationEvent) error
}

// PriorityBoostAmountCents returns the fee for a priority boost join.
// Uses PRIORITY_BOOST_AMOUNT_CENTS env (default 100). If boost level > 1, uses that as cents.
func PriorityBoostAmountCents(boostLevel int32) int64 {
	if boostLevel > 1 {
		return int64(boostLevel)
	}
	if v := os.Getenv("PRIORITY_BOOST_AMOUNT_CENTS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n >= 0 {
			return n
		}
	}
	return 100
}

// BuildQueuePriorityBoostEvent builds a ChargeableOperationRequested for priority boost.
func BuildQueuePriorityBoostEvent(
	playerID, gameID, region, tenantID, clientID, resourceOwnerID, correlationID string,
	boostLevel int32,
) *schemas.ChargeableOperationEvent {
	now := time.Now().UTC()
	amount := PriorityBoostAmountCents(boostLevel)
	idem := fmt.Sprintf("%s:%s:%s:%s",
		schemas.OperationTypeQueuePriorityBoost, playerID, gameID, correlationID)
	if correlationID == "" {
		idem = fmt.Sprintf("%s:%s:%s:%d",
			schemas.OperationTypeQueuePriorityBoost, playerID, gameID, now.UnixMilli())
	}

	return &schemas.ChargeableOperationEvent{
		ID:                uuid.New().String(),
		Type:              schemas.EventTypeChargeableOperationRequested,
		Source:            "match-making-api",
		SpecVersion:       schemas.CloudEventsSpecVersion,
		TimeUnixMs:        now.UnixMilli(),
		Subject:           playerID,
		ResourceOwnerID:   resourceOwnerID,
		CorrelationID:     correlationID,
		DataschemaVersion: schemas.SchemaVersionV1,
		ChargeableOperationRequested: &schemas.ChargeableOperationRequestedPayload{
			OperationType:   schemas.OperationTypeQueuePriorityBoost,
			AmountCents:     amount,
			Currency:        "USD",
			PlayerID:        playerID,
			ResourceOwnerID: resourceOwnerID,
			TenantID:        tenantID,
			ClientID:        clientID,
			GameID:          gameID,
			Region:          region,
			CorrelationID:   correlationID,
			IdempotencyKey:  idem,
			RequestedAtMs:   now.UnixMilli(),
		},
	}
}

// publishChargeableIfPriorityBoost publishes ChargeableOperationRequested when boost > 0.
// Failures are logged and non-fatal so queue join is not blocked by wallet topic outages.
func (c *MatchmakingEventConsumer) publishChargeableIfPriorityBoost(
	ctx context.Context,
	envelope *schemas.EventEnvelope,
	payload *schemas.PlayerQueuedPayload,
	playerID uuid.UUID,
) {
	pb := payload.PriorityBoost
	if pb == nil || *pb <= 0 {
		return
	}
	pub, ok := c.eventPublisher.(ChargeablePublisher)
	if !ok || pub == nil {
		slog.WarnContext(ctx, "ChargeablePublisher not available; skipping ChargeableOperationRequested",
			"player_id", playerID)
		return
	}

	event := BuildQueuePriorityBoostEvent(
		playerID.String(),
		payload.GetGameId(),
		payload.GetRegion(),
		payload.GetTenantId(),
		payload.GetClientId(),
		envelope.GetResourceOwnerId(),
		envelope.GetCorrelationId(),
		*pb,
	)
	if err := pub.PublishChargeableOperationRequested(ctx, event); err != nil {
		slog.ErrorContext(ctx, "Failed to publish ChargeableOperationRequested",
			"error", err,
			"player_id", playerID,
			"operation_type", schemas.OperationTypeQueuePriorityBoost)
		return
	}
	slog.InfoContext(ctx, "Published ChargeableOperationRequested for priority boost",
		"player_id", playerID,
		"amount_cents", event.ChargeableOperationRequested.AmountCents,
		"idempotency_key", event.ChargeableOperationRequested.IdempotencyKey)
}
