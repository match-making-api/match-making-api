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

// MatchStartedHandler is the domain-level function that processes a validated MatchStarted event.
type MatchStartedHandler func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.MatchStartedPayload) error

// MatchStartedConsumer consumes MatchStarted events from matchmaking.match.started.
// Validates resource ownership and delegates to the domain handler for WebSocket broadcast.
type MatchStartedConsumer struct {
	consumer *Consumer
	handler  MatchStartedHandler
}

// NewMatchStartedConsumer creates a consumer for the matchmaking.match.started topic.
func NewMatchStartedConsumer(client *Client, groupID string, handler MatchStartedHandler) *MatchStartedConsumer {
	config := DefaultConsumerConfig(groupID, []string{TopicMatchStarted})
	consumer := NewConsumer(client, config)

	msc := &MatchStartedConsumer{
		consumer: consumer,
		handler:  handler,
	}

	consumer.RegisterHandler(TopicMatchStarted, msc.handleMessage)

	return msc
}

// handleMessage deserializes and routes a single Kafka message from matchmaking.match.started.
func (msc *MatchStartedConsumer) handleMessage(ctx context.Context, msg *kafkago.Message) error {
	var event schemas.MatchmakingEvent
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal MatchmakingEvent (MatchStarted)",
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

	data, ok := event.GetData().(*schemas.MatchmakingEvent_MatchStarted)
	if !ok || data.MatchStarted == nil {
		slog.WarnContext(ctx, "Unknown or missing MatchStarted payload, skipping",
			"event_id", envelope.GetId(),
			"event_type", envelope.GetType())
		return nil
	}

	if err := validateMatchStartedResourceOwnership(envelope, data.MatchStarted); err != nil {
		slog.ErrorContext(ctx, "MatchStarted resource ownership validation failed, skipping",
			"event_id", envelope.GetId(),
			"match_id", data.MatchStarted.GetMatchId(),
			"error", err)
		return nil
	}

	slog.InfoContext(ctx, "Processing MatchStarted event",
		"event_id", envelope.GetId(),
		"match_id", data.MatchStarted.GetMatchId(),
		"resource_owner_id", envelope.GetResourceOwnerId())

	return msc.handler(ctx, envelope, data.MatchStarted)
}

// validateMatchStartedResourceOwnership checks required fields for match participant delivery.
func validateMatchStartedResourceOwnership(envelope *schemas.EventEnvelope, payload *schemas.MatchStartedPayload) error {
	if strings.TrimSpace(envelope.GetResourceOwnerId()) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty in envelope", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetMatchId()) == "" {
		return fmt.Errorf("%w: match_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetResourceOwnerId()) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if len(payload.GetPlayerIds()) == 0 {
		return fmt.Errorf("%w: player_ids is empty (required for MatchStarted broadcast)", ErrResourceOwnershipInvalid)
	}
	return nil
}

// Start begins consuming messages from matchmaking.match.started.
func (msc *MatchStartedConsumer) Start(ctx context.Context) error {
	return msc.consumer.Start(ctx)
}

// Close closes the consumer.
func (msc *MatchStartedConsumer) Close() error {
	return msc.consumer.Close()
}
