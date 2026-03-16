package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/segmentio/kafka-go"
)

// ReadyCheckConsumerConfig holds configuration and dependencies for the ready check consumer
type ReadyCheckConsumerConfig struct {
	GroupID           string
	CommitmentWriter  pairing_out.CommitmentWriter
	CommitmentReader  pairing_out.CommitmentReader
	EventPublisher    *EventPublisher
	NotificationFunc  func(ctx context.Context, lobbyID uuid.UUID, playerIDs []uuid.UUID, notifType pairing_entities.NotificationType, title, message string, metadata map[string]interface{})
}

// ReadyCheckConsumer consumes ready check lifecycle events from Kafka
type ReadyCheckConsumer struct {
	consumer       *Consumer
	config         *ReadyCheckConsumerConfig
}

// NewReadyCheckConsumer creates a new ready check event consumer
func NewReadyCheckConsumer(client *Client, config *ReadyCheckConsumerConfig) *ReadyCheckConsumer {
	consumerConfig := DefaultConsumerConfig(config.GroupID, []string{TopicReadyCheck})
	consumer := NewConsumer(client, consumerConfig)

	rc := &ReadyCheckConsumer{
		consumer: consumer,
		config:   config,
	}

	consumer.RegisterHandler(TopicReadyCheck, rc.handleReadyCheckEvent)

	return rc
}

func (rc *ReadyCheckConsumer) handleReadyCheckEvent(ctx context.Context, msg *kafka.Message) error {
	var event ReadyCheckEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal ready check event: %w", err)
	}

	slog.InfoContext(ctx, "Processing ready check event",
		"event_type", event.EventType,
		"lobby_id", event.LobbyID,
		"player_id", event.PlayerID)

	switch event.EventType {
	case EventTypeReadyCheckStarted:
		return rc.handleReadyCheckStarted(ctx, &event)
	case EventTypeReadinessConfirmed:
		return rc.handleReadinessConfirmed(ctx, &event)
	case EventTypeReadinessDeclined:
		return rc.handleReadinessDeclined(ctx, &event)
	case EventTypeReadyCheckTimeout:
		return rc.handleReadyCheckTimeout(ctx, &event)
	case EventTypeAllPlayersReady:
		return rc.handleAllPlayersReady(ctx, &event)
	case EventTypeGameConnectionDelivered:
		return rc.handleGameConnectionDelivered(ctx, &event)
	default:
		slog.WarnContext(ctx, "Unknown ready check event type", "event_type", event.EventType)
		return nil
	}
}

func (rc *ReadyCheckConsumer) handleReadyCheckStarted(ctx context.Context, event *ReadyCheckEvent) error {
	slog.InfoContext(ctx, "Ready check started",
		"lobby_id", event.LobbyID,
		"player_count", len(event.PlayerIDs))

	// Broadcast via WebSocket (delegated to bridge)
	// Additional processing: metric tracking, audit log
	return nil
}

func (rc *ReadyCheckConsumer) handleReadinessConfirmed(ctx context.Context, event *ReadyCheckEvent) error {
	if event.PlayerID == nil {
		return fmt.Errorf("readiness_confirmed event missing player_id")
	}

	slog.InfoContext(ctx, "Player confirmed readiness",
		"lobby_id", event.LobbyID,
		"player_id", *event.PlayerID)

	return nil
}

func (rc *ReadyCheckConsumer) handleReadinessDeclined(ctx context.Context, event *ReadyCheckEvent) error {
	if event.PlayerID == nil {
		return fmt.Errorf("readiness_declined event missing player_id")
	}

	slog.InfoContext(ctx, "Player declined readiness",
		"lobby_id", event.LobbyID,
		"player_id", *event.PlayerID)

	return nil
}

func (rc *ReadyCheckConsumer) handleReadyCheckTimeout(ctx context.Context, event *ReadyCheckEvent) error {
	slog.InfoContext(ctx, "Ready check timed out",
		"lobby_id", event.LobbyID)

	return nil
}

func (rc *ReadyCheckConsumer) handleAllPlayersReady(ctx context.Context, event *ReadyCheckEvent) error {
	slog.InfoContext(ctx, "All players ready",
		"lobby_id", event.LobbyID,
		"player_count", len(event.PlayerIDs))

	return nil
}

func (rc *ReadyCheckConsumer) handleGameConnectionDelivered(ctx context.Context, event *ReadyCheckEvent) error {
	slog.InfoContext(ctx, "Game connection info delivered",
		"lobby_id", event.LobbyID)

	return nil
}

// Start begins consuming ready check events
func (rc *ReadyCheckConsumer) Start(ctx context.Context) error {
	slog.Info("Starting ready check consumer", "group_id", rc.config.GroupID)
	return rc.consumer.Start(ctx)
}

// Close shuts down the consumer
func (rc *ReadyCheckConsumer) Close() error {
	return rc.consumer.Close()
}
