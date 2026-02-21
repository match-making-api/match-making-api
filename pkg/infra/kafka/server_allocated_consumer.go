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

// ServerAllocatedHandler is the domain-level function that processes a validated ServerAllocated event.
type ServerAllocatedHandler func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.ServerAllocatedPayload) error

// ServerAllocatedConsumer consumes ServerAllocated events from matchmaking.server.allocated.
// Validates resource ownership and delegates to the domain handler for MatchReady broadcast.
type ServerAllocatedConsumer struct {
	consumer *Consumer
	handler  ServerAllocatedHandler
}

// NewServerAllocatedConsumer creates a consumer for the matchmaking.server.allocated topic.
func NewServerAllocatedConsumer(client *Client, groupID string, handler ServerAllocatedHandler) *ServerAllocatedConsumer {
	config := DefaultConsumerConfig(groupID, []string{TopicServerAllocated})
	consumer := NewConsumer(client, config)

	sac := &ServerAllocatedConsumer{
		consumer: consumer,
		handler:  handler,
	}

	consumer.RegisterHandler(TopicServerAllocated, sac.handleMessage)

	return sac
}

// handleMessage deserializes and routes a single Kafka message from matchmaking.server.allocated.
func (sac *ServerAllocatedConsumer) handleMessage(ctx context.Context, msg *kafkago.Message) error {
	var event schemas.MatchmakingEvent
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal MatchmakingEvent (ServerAllocated)",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"error", err)
		return nil // Skip malformed messages
	}

	envelope := event.GetEnvelope()
	if envelope == nil {
		slog.ErrorContext(ctx, "MatchmakingEvent has nil envelope, skipping",
			"topic", msg.Topic,
			"offset", msg.Offset)
		return nil
	}

	data, ok := event.GetData().(*schemas.MatchmakingEvent_ServerAllocated)
	if !ok || data.ServerAllocated == nil {
		slog.WarnContext(ctx, "Unknown or missing ServerAllocated payload, skipping",
			"event_id", envelope.GetId(),
			"event_type", envelope.GetType())
		return nil
	}

	if err := validateServerAllocatedResourceOwnership(envelope, data.ServerAllocated); err != nil {
		slog.ErrorContext(ctx, "ServerAllocated resource ownership validation failed, skipping",
			"event_id", envelope.GetId(),
			"match_id", data.ServerAllocated.GetMatchId(),
			"error", err)
		return nil
	}

	slog.InfoContext(ctx, "Processing ServerAllocated event",
		"event_id", envelope.GetId(),
		"match_id", data.ServerAllocated.GetMatchId(),
		"server_id", data.ServerAllocated.GetServerId(),
		"resource_owner_id", envelope.GetResourceOwnerId())

	return sac.handler(ctx, envelope, data.ServerAllocated)
}

// validateServerAllocatedResourceOwnership checks required fields for server access control.
func validateServerAllocatedResourceOwnership(envelope *schemas.EventEnvelope, payload *schemas.ServerAllocatedPayload) error {
	if strings.TrimSpace(envelope.GetResourceOwnerId()) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty in envelope", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetMatchId()) == "" {
		return fmt.Errorf("%w: match_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetServerId()) == "" {
		return fmt.Errorf("%w: server_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetResourceOwnerId()) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if len(payload.GetPlayerIds()) == 0 {
		return fmt.Errorf("%w: player_ids is empty (required for MatchReady broadcast)", ErrResourceOwnershipInvalid)
	}
	return nil
}

// Start begins consuming messages from matchmaking.server.allocated.
func (sac *ServerAllocatedConsumer) Start(ctx context.Context) error {
	return sac.consumer.Start(ctx)
}

// Close closes the consumer.
func (sac *ServerAllocatedConsumer) Close() error {
	return sac.consumer.Close()
}
