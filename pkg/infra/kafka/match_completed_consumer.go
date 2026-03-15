package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	kafkago "github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// MatchCompletedHandler is the domain-level function that processes a validated MatchCompleted event.
type MatchCompletedHandler func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.MatchCompletedPayload) error

// MatchCompletedConsumer consumes MatchCompleted events from matchmaking.matches.completed.
// Validates resource ownership and delegates to the domain handler.
type MatchCompletedConsumer struct {
	consumer *Consumer
	handler  MatchCompletedHandler
}

// NewMatchCompletedConsumer creates a consumer for the matchmaking.matches.completed topic.
func NewMatchCompletedConsumer(client *Client, groupID string, handler MatchCompletedHandler) *MatchCompletedConsumer {
	config := DefaultConsumerConfig(groupID, []string{TopicMatchCompleted})
	consumer := NewConsumer(client, config)

	mcc := &MatchCompletedConsumer{
		consumer: consumer,
		handler:  handler,
	}

	consumer.RegisterHandler(TopicMatchCompleted, mcc.handleMessage)

	return mcc
}

// handleMessage deserializes and routes a single Kafka message from matchmaking.matches.completed.
func (mcc *MatchCompletedConsumer) handleMessage(ctx context.Context, msg *kafkago.Message) error {
	var event schemas.MatchmakingEvent
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal MatchmakingEvent (MatchCompleted)",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"error", err)
		return nil
	}

	envelope := event.GetEnvelope()
	if envelope == nil {
		slog.ErrorContext(ctx, "MatchmakingEvent has nil envelope, skipping",
			"topic", msg.Topic,
			"offset", msg.Offset)
		return nil
	}

	data, ok := event.GetData().(*schemas.MatchmakingEvent_MatchCompleted)
	if !ok || data.MatchCompleted == nil {
		slog.WarnContext(ctx, "Unknown or missing MatchCompleted payload, skipping",
			"event_id", envelope.GetId(),
			"event_type", envelope.GetType())
		return nil
	}

	if err := validateMatchCompletedResourceOwnership(envelope, data.MatchCompleted); err != nil {
		slog.ErrorContext(ctx, "MatchCompleted resource ownership validation failed, skipping",
			"event_id", envelope.GetId(),
			"match_id", data.MatchCompleted.GetMatchId(),
			"error", err)
		return nil
	}

	slog.InfoContext(ctx, "Processing MatchCompleted event",
		"event_id", envelope.GetId(),
		"match_id", data.MatchCompleted.GetMatchId(),
		"resource_owner_id", envelope.GetResourceOwnerId())

	return mcc.handler(ctx, envelope, data.MatchCompleted)
}

// validateMatchCompletedResourceOwnership checks required fields for results processing.
func validateMatchCompletedResourceOwnership(envelope *schemas.EventEnvelope, payload *schemas.MatchCompletedPayload) error {
	if strings.TrimSpace(envelope.GetResourceOwnerId()) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty in envelope", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetMatchId()) == "" {
		return fmt.Errorf("%w: match_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetTenantId()) == "" {
		return fmt.Errorf("%w: tenant_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetClientId()) == "" {
		return fmt.Errorf("%w: client_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	return nil
}

// Start begins consuming messages from matchmaking.matches.completed.
func (mcc *MatchCompletedConsumer) Start(ctx context.Context) error {
	return mcc.consumer.Start(ctx)
}

// Close closes the consumer.
func (mcc *MatchCompletedConsumer) Close() error {
	return mcc.consumer.Close()
}
