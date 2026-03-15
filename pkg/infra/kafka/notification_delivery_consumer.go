package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/segmentio/kafka-go"
)

// NotificationDeliveryEvent represents a notification delivery request from Kafka
type NotificationDeliveryEvent struct {
	EventID        uuid.UUID                              `json:"event_id"`
	UserID         uuid.UUID                              `json:"user_id"`
	Channel        pairing_entities.NotificationChannel   `json:"channel"`
	Type           pairing_entities.NotificationType      `json:"type"`
	Title          string                                 `json:"title"`
	Message        string                                 `json:"message"`
	Metadata       map[string]interface{}                 `json:"metadata,omitempty"`
	TemplateID     *uuid.UUID                             `json:"template_id,omitempty"`
	Language       string                                 `json:"language,omitempty"`
	MaxRetries     int                                    `json:"max_retries,omitempty"`
	Timestamp      int64                                  `json:"timestamp"`
	RetryCount     int                                    `json:"retry_count,omitempty"`
}

// NotificationDeliveryConsumerConfig holds dependencies for the notification delivery consumer
type NotificationDeliveryConsumerConfig struct {
	GroupID              string
	NotificationWriter   pairing_out.NotificationWriter
	PreferencesReader    pairing_out.UserNotificationPreferencesReader
	SenderFactory        pairing_out.NotificationSenderResolver
	EventPublisher       *EventPublisher
}

// NotificationDeliveryConsumer consumes notification delivery requests from Kafka
// and dispatches them to the appropriate channel sender
type NotificationDeliveryConsumer struct {
	consumer    *Consumer
	config      *NotificationDeliveryConsumerConfig
}

// NewNotificationDeliveryConsumer creates a new notification delivery consumer
func NewNotificationDeliveryConsumer(client *Client, config *NotificationDeliveryConsumerConfig) *NotificationDeliveryConsumer {
	consumerConfig := DefaultConsumerConfig(config.GroupID, []string{TopicNotificationDelivery})
	consumer := NewConsumer(client, consumerConfig)

	ndc := &NotificationDeliveryConsumer{
		consumer: consumer,
		config:   config,
	}

	consumer.RegisterHandler(TopicNotificationDelivery, ndc.handleNotificationDelivery)

	return ndc
}

func (ndc *NotificationDeliveryConsumer) handleNotificationDelivery(ctx context.Context, msg *kafka.Message) error {
	var event NotificationDeliveryEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal notification delivery event: %w", err)
	}

	slog.InfoContext(ctx, "Processing notification delivery",
		"event_id", event.EventID,
		"user_id", event.UserID,
		"channel", event.Channel,
		"type", event.Type,
		"retry_count", event.RetryCount)

	// Get sender for channel
	sender := ndc.config.SenderFactory.GetSender(event.Channel)
	if sender == nil || !sender.IsAvailable(ctx) {
		// Fallback to in-app notification
		slog.WarnContext(ctx, "Sender not available, falling back to in-app",
			"channel", event.Channel,
			"user_id", event.UserID)
		sender = ndc.config.SenderFactory.GetSender(pairing_entities.NotificationChannelInApp)
	}

	if sender == nil {
		return ndc.sendToDLQ(ctx, &event, "no sender available for any channel")
	}

	// Check user preferences
	preferences, err := ndc.config.PreferencesReader.GetByUserID(ctx, event.UserID)
	if err != nil {
		// Use defaults if no preferences found
		slog.DebugContext(ctx, "No preferences found, using defaults", "user_id", event.UserID)
	}

	if preferences != nil && !preferences.IsChannelEnabled(event.Channel) {
		slog.InfoContext(ctx, "Channel disabled for user, skipping",
			"channel", event.Channel,
			"user_id", event.UserID)
		return nil // Silently skip
	}

	// Determine language
	language := event.Language
	if language == "" && preferences != nil {
		language = preferences.PreferredLanguage
	}
	if language == "" {
		language = "en"
	}

	maxRetries := event.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	// Create notification entity
	notification := pairing_entities.NewNotification(
		common.ResourceOwner{UserID: event.UserID}, // resource owner from event
		event.UserID,
		event.Channel,
		event.Type,
		event.Title,
		event.Message,
		event.Metadata,
		language,
		maxRetries,
		nil,
	)

	if event.TemplateID != nil {
		notification.TemplateID = event.TemplateID
	}

	// Save notification
	savedNotification, err := ndc.config.NotificationWriter.Save(ctx, notification)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save notification", "error", err)
		return ndc.sendToDLQ(ctx, &event, fmt.Sprintf("save failed: %v", err))
	}

	// Send notification via channel sender
	if err := sender.Send(ctx, savedNotification); err != nil {
		slog.ErrorContext(ctx, "Failed to send notification",
			"error", err,
			"notification_id", savedNotification.ID,
			"channel", event.Channel)

		savedNotification.MarkAsFailed(err.Error())
		ndc.config.NotificationWriter.Save(ctx, savedNotification)

		// Retry logic
		if event.RetryCount < maxRetries {
			return ndc.retryDelivery(ctx, &event)
		}

		return ndc.sendToDLQ(ctx, &event, fmt.Sprintf("max retries exceeded: %v", err))
	}

	slog.InfoContext(ctx, "Notification delivered successfully",
		"notification_id", savedNotification.ID,
		"channel", event.Channel,
		"user_id", event.UserID)

	return nil
}

func (ndc *NotificationDeliveryConsumer) retryDelivery(ctx context.Context, event *NotificationDeliveryEvent) error {
	event.RetryCount++
	event.Timestamp = time.Now().UnixMilli()

	msg := &Message{
		Key:       event.UserID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"retry_count": fmt.Sprintf("%d", event.RetryCount),
			"event_type":  string(event.Type),
		},
	}

	slog.InfoContext(ctx, "Retrying notification delivery",
		"event_id", event.EventID,
		"retry_count", event.RetryCount)

	return ndc.config.EventPublisher.client.Publish(ctx, TopicNotificationDelivery, msg)
}

func (ndc *NotificationDeliveryConsumer) sendToDLQ(ctx context.Context, event *NotificationDeliveryEvent, reason string) error {
	dlqEvent := map[string]interface{}{
		"original_event": event,
		"reason":        reason,
		"timestamp":     time.Now().UnixMilli(),
	}

	msg := &Message{
		Key:       event.UserID.String(),
		Value:     dlqEvent,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"reason":     reason,
			"event_type": string(event.Type),
		},
	}

	slog.WarnContext(ctx, "Sending notification to DLQ",
		"event_id", event.EventID,
		"reason", reason)

	return ndc.config.EventPublisher.client.Publish(ctx, TopicNotificationDeliveryDLQ, msg)
}

// Start begins consuming notification delivery events
func (ndc *NotificationDeliveryConsumer) Start(ctx context.Context) error {
	slog.Info("Starting notification delivery consumer", "group_id", ndc.config.GroupID)
	return ndc.consumer.Start(ctx)
}

// Close shuts down the consumer
func (ndc *NotificationDeliveryConsumer) Close() error {
	return ndc.consumer.Close()
}
