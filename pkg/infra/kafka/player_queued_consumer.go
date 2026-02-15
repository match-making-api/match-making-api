package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	kafkago "github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// ErrResourceOwnershipInvalid is returned when resource ownership validation fails on the consumer side.
var ErrResourceOwnershipInvalid = errors.New("resource ownership validation failed")

// PlayerQueuedHandler is the domain-level function that processes a validated PlayerQueued event.
// It receives the envelope (with resource ownership metadata) and the typed payload.
type PlayerQueuedHandler func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error

// PlayerQueuedConsumer consumes PlayerQueued events from the matchmaking.commands topic.
// It deserializes protobuf/CloudEvents messages, validates resource ownership, and delegates
// to a domain handler for pool management.
type PlayerQueuedConsumer struct {
	consumer *Consumer
	handler  PlayerQueuedHandler
}

// NewPlayerQueuedConsumer creates a consumer for the matchmaking.commands topic.
// The handler function is called for each valid PlayerQueued event after resource ownership validation.
func NewPlayerQueuedConsumer(client *Client, groupID string, handler PlayerQueuedHandler) *PlayerQueuedConsumer {
	config := DefaultConsumerConfig(groupID, []string{TopicMatchmakingCommands})
	consumer := NewConsumer(client, config)

	pqc := &PlayerQueuedConsumer{
		consumer: consumer,
		handler:  handler,
	}

	consumer.RegisterHandler(TopicMatchmakingCommands, pqc.handleMessage)

	return pqc
}

// handleMessage deserializes and routes a single Kafka message from matchmaking.commands.
func (pqc *PlayerQueuedConsumer) handleMessage(ctx context.Context, msg *kafkago.Message) error {
	// Deserialize MatchmakingEvent from protojson
	var event schemas.MatchmakingEvent
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal MatchmakingEvent",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"error", err)
		// Return nil to skip (commit) malformed messages — they can't be retried
		return nil
	}

	envelope := event.GetEnvelope()
	if envelope == nil {
		slog.ErrorContext(ctx, "MatchmakingEvent has nil envelope, skipping",
			"topic", msg.Topic,
			"offset", msg.Offset)
		return nil
	}

	slog.InfoContext(ctx, "Received matchmaking command",
		"event_id", envelope.GetId(),
		"event_type", envelope.GetType(),
		"source", envelope.GetSource(),
		"subject", envelope.GetSubject())

	// Route based on event type / oneof data
	switch data := event.GetData().(type) {
	case *schemas.MatchmakingEvent_PlayerQueued:
		return pqc.handlePlayerQueued(ctx, envelope, data.PlayerQueued)
	default:
		slog.WarnContext(ctx, "Unknown or unsupported event type in matchmaking.commands, skipping",
			"event_type", envelope.GetType(),
			"event_id", envelope.GetId())
		return nil
	}
}

// handlePlayerQueued validates resource ownership and delegates to the domain handler.
func (pqc *PlayerQueuedConsumer) handlePlayerQueued(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
	// Validate resource ownership from envelope (Epic §9)
	if err := validateResourceOwnership(envelope, payload); err != nil {
		slog.ErrorContext(ctx, "PlayerQueued resource ownership validation failed, skipping",
			"event_id", envelope.GetId(),
			"player_id", payload.GetPlayerId(),
			"error", err)
		// Skip invalid events — don't return error to avoid reprocessing
		return nil
	}

	slog.InfoContext(ctx, "Processing PlayerQueued event",
		"event_id", envelope.GetId(),
		"player_id", payload.GetPlayerId(),
		"game_id", payload.GetGameId(),
		"region", payload.GetRegion(),
		"resource_owner_id", envelope.GetResourceOwnerId())

	return pqc.handler(ctx, envelope, payload)
}

// validateResourceOwnership checks that required multi-tenancy fields are present.
func validateResourceOwnership(envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
	if strings.TrimSpace(envelope.GetResourceOwnerId()) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty in envelope", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetTenantId()) == "" {
		return fmt.Errorf("%w: tenant_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetClientId()) == "" {
		return fmt.Errorf("%w: client_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetPlayerId()) == "" {
		return fmt.Errorf("%w: player_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	return nil
}

// Start begins consuming messages from matchmaking.commands.
func (pqc *PlayerQueuedConsumer) Start(ctx context.Context) error {
	return pqc.consumer.Start(ctx)
}

// Close closes the consumer.
func (pqc *PlayerQueuedConsumer) Close() error {
	return pqc.consumer.Close()
}
