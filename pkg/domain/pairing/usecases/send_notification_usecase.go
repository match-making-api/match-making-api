package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// SendNotificationUseCase handles sending a single notification
type SendNotificationUseCase struct {
	NotificationWriter              pairing_out.NotificationWriter
	NotificationReader              pairing_out.NotificationReader
	UserNotificationPreferencesReader pairing_out.UserNotificationPreferencesReader
	SenderFactory                   *NotificationSenderFactory
}

// SendNotificationPayload contains the information needed to send a notification
type SendNotificationPayload struct {
	UserID      uuid.UUID
	Channel     pairing_entities.NotificationChannel
	Type        pairing_entities.NotificationType
	Title       string
	Message     string
	Metadata    map[string]interface{}
	TemplateID  *uuid.UUID
	Language    string
	MaxRetries  int
}

// Execute sends a notification after validating user preferences and channel availability
func (uc *SendNotificationUseCase) Execute(ctx context.Context, payload SendNotificationPayload) (*pairing_entities.Notification, error) {
	resourceOwner := common.GetResourceOwner(ctx)
	
	// Get user preferences
	preferences, err := uc.UserNotificationPreferencesReader.GetByUserID(ctx, payload.UserID)
	if err != nil {
		// If preferences don't exist, use defaults
		preferences = pairing_entities.NewUserNotificationPreferences(resourceOwner, payload.UserID, payload.Language)
	}

	// Check if channel is enabled for user
	if !preferences.IsChannelEnabled(payload.Channel) {
		return nil, fmt.Errorf("channel %v is disabled for user %v", payload.Channel, payload.UserID)
	}

	// Check if notification type is enabled for user
	if !preferences.IsTypeEnabled(payload.Type) {
		return nil, fmt.Errorf("notification type %v is disabled for user %v", payload.Type, payload.UserID)
	}

	// Check do not disturb time
	if err := uc.checkDoNotDisturb(ctx, preferences); err != nil {
		return nil, fmt.Errorf("cannot send notification during do not disturb time: %w", err)
	}

	// Get sender for channel
	sender := uc.SenderFactory.GetSender(payload.Channel)
	if sender == nil {
		return nil, fmt.Errorf("no sender available for channel %v", payload.Channel)
	}

	// Check if sender is available
	if !sender.IsAvailable(ctx) {
		return nil, fmt.Errorf("sender for channel %v is not available", payload.Channel)
	}

	// Determine language
	language := payload.Language
	if language == "" {
		language = preferences.PreferredLanguage
	}
	if language == "" {
		language = "en" // Default to English
	}

	// Determine max retries
	maxRetries := payload.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3 // Default to 3 retries
	}

	// Get created by (if admin)
	var createdBy *uuid.UUID
	if common.IsAdmin(ctx) {
		userID := common.GetResourceOwner(ctx).UserID
		createdBy = &userID
	}

	// Create notification
	notification := pairing_entities.NewNotification(
		resourceOwner,
		payload.UserID,
		payload.Channel,
		payload.Type,
		payload.Title,
		payload.Message,
		payload.Metadata,
		language,
		maxRetries,
		createdBy,
	)

	if payload.TemplateID != nil {
		notification.TemplateID = payload.TemplateID
	}

	// Save notification
	savedNotification, err := uc.NotificationWriter.Save(ctx, notification)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save notification", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification
	if err := sender.Send(ctx, savedNotification); err != nil {
		slog.ErrorContext(ctx, "failed to send notification", "error", err, "notification_id", savedNotification.ID)
		savedNotification.MarkAsFailed(err.Error())
		_, saveErr := uc.NotificationWriter.Save(ctx, savedNotification)
		if saveErr != nil {
			slog.ErrorContext(ctx, "failed to save failed notification", "error", saveErr)
		}
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	// Update notification status
	savedNotification, err = uc.NotificationWriter.Save(ctx, savedNotification)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification status", "error", err)
		// Don't fail the operation if status update fails
	}

	slog.InfoContext(ctx, "notification sent successfully",
		"notification_id", savedNotification.ID,
		"user_id", payload.UserID,
		"channel", payload.Channel,
		"type", payload.Type)

	return savedNotification, nil
}

// checkDoNotDisturb checks if current time is within do not disturb period
func (uc *SendNotificationUseCase) checkDoNotDisturb(ctx context.Context, preferences *pairing_entities.UserNotificationPreferences) error {
	if preferences.DoNotDisturbStart == nil || preferences.DoNotDisturbEnd == nil {
		return nil // No DND configured
	}

	now := time.Now()
	currentTime := now.Format("15:04") // HH:MM format

	start := *preferences.DoNotDisturbStart
	end := *preferences.DoNotDisturbEnd

	// Simple time comparison (assumes same day)
	if currentTime >= start && currentTime <= end {
		return fmt.Errorf("current time %s is within do not disturb period (%s - %s)", currentTime, start, end)
	}

	return nil
}
