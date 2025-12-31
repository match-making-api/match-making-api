package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// SendBatchNotificationUseCase handles sending notifications to multiple users
type SendBatchNotificationUseCase struct {
	NotificationWriter              pairing_out.NotificationWriter
	UserNotificationPreferencesReader pairing_out.UserNotificationPreferencesReader
	SenderFactory                   *NotificationSenderFactory
}

// SendBatchNotificationPayload contains the information needed to send batch notifications
type SendBatchNotificationPayload struct {
	UserIDs     []uuid.UUID
	Channel     pairing_entities.NotificationChannel
	Type        pairing_entities.NotificationType
	Title       string
	Message     string
	Metadata    map[string]interface{}
	TemplateID  *uuid.UUID
	Language    string
	MaxRetries  int
}

// Execute sends notifications to multiple users
func (uc *SendBatchNotificationUseCase) Execute(ctx context.Context, payload SendBatchNotificationPayload) ([]*pairing_entities.Notification, []error) {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return nil, []error{fmt.Errorf("only administrators can send batch notifications")}
	}

	if len(payload.UserIDs) == 0 {
		return nil, []error{fmt.Errorf("user_ids list cannot be empty")}
	}

	resourceOwner := common.GetResourceOwner(ctx)
	var createdBy *uuid.UUID
	if common.IsAdmin(ctx) {
		userID := resourceOwner.UserID
		createdBy = &userID
	}

	var notifications []*pairing_entities.Notification
	var errors []error

	// Get sender for channel
	sender := uc.SenderFactory.GetSender(payload.Channel)
	if sender == nil {
		return nil, []error{fmt.Errorf("no sender available for channel %v", payload.Channel)}
	}

	if !sender.IsAvailable(ctx) {
		return nil, []error{fmt.Errorf("sender for channel %v is not available", payload.Channel)}
	}

	// Process each user
	for _, userID := range payload.UserIDs {
		// Get user preferences
		preferences, err := uc.UserNotificationPreferencesReader.GetByUserID(ctx, userID)
		if err != nil {
			// If preferences don't exist, use defaults
			preferences = pairing_entities.NewUserNotificationPreferences(resourceOwner, userID, payload.Language)
		}

		// Check if channel is enabled for user
		if !preferences.IsChannelEnabled(payload.Channel) {
			errors = append(errors, fmt.Errorf("channel disabled for user %v", userID))
			continue
		}

		// Check if notification type is enabled for user
		if !preferences.IsTypeEnabled(payload.Type) {
			errors = append(errors, fmt.Errorf("notification type disabled for user %v", userID))
			continue
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

		// Create notification
		notification := pairing_entities.NewNotification(
			resourceOwner,
			userID,
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
			slog.ErrorContext(ctx, "failed to save notification in batch", "error", err, "user_id", userID)
			errors = append(errors, fmt.Errorf("failed to create notification for user %v: %w", userID, err))
			continue
		}

		// Send notification
		if err := sender.Send(ctx, savedNotification); err != nil {
			slog.ErrorContext(ctx, "failed to send notification in batch", "error", err, "notification_id", savedNotification.ID)
			savedNotification.MarkAsFailed(err.Error())
			_, saveErr := uc.NotificationWriter.Save(ctx, savedNotification)
			if saveErr != nil {
				slog.ErrorContext(ctx, "failed to save failed notification", "error", saveErr)
			}
			errors = append(errors, fmt.Errorf("failed to send notification for user %v: %w", userID, err))
			continue
		}

		// Update notification status
		savedNotification, err = uc.NotificationWriter.Save(ctx, savedNotification)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update notification status", "error", err)
		}

		notifications = append(notifications, savedNotification)
	}

	slog.InfoContext(ctx, "batch notification sent",
		"total_users", len(payload.UserIDs),
		"successful", len(notifications),
		"failed", len(errors),
		"channel", payload.Channel,
		"type", payload.Type)

	return notifications, errors
}
