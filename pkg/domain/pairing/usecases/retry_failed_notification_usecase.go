package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// RetryFailedNotificationUseCase handles retrying failed notifications
type RetryFailedNotificationUseCase struct {
	NotificationReader pairing_out.NotificationReader
	NotificationWriter  pairing_out.NotificationWriter
	SenderFactory       *NotificationSenderFactory
	RetryDelay          time.Duration // Delay between retries
}

// NewRetryFailedNotificationUseCase creates a new retry use case
func NewRetryFailedNotificationUseCase(
	notificationReader pairing_out.NotificationReader,
	notificationWriter pairing_out.NotificationWriter,
	senderFactory *NotificationSenderFactory,
	retryDelay time.Duration,
) *RetryFailedNotificationUseCase {
	if retryDelay == 0 {
		retryDelay = 5 * time.Minute // Default 5 minutes
	}
	return &RetryFailedNotificationUseCase{
		NotificationReader: notificationReader,
		NotificationWriter:  notificationWriter,
		SenderFactory:       senderFactory,
		RetryDelay:        retryDelay,
	}
}

// Execute retries a specific failed notification
func (uc *RetryFailedNotificationUseCase) Execute(ctx context.Context, notificationID uuid.UUID) error {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return fmt.Errorf("only administrators can retry notifications")
	}

	// Get notification
	notification, err := uc.NotificationReader.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	// Check if notification can be retried
	if !notification.CanRetry() {
		return fmt.Errorf("notification cannot be retried (status: %v, retry_count: %d, max_retries: %d)",
			notification.Status, notification.RetryCount, notification.MaxRetries)
	}

	// Get sender for channel
	sender := uc.SenderFactory.GetSender(notification.Channel)
	if sender == nil {
		return fmt.Errorf("no sender available for channel %v", notification.Channel)
	}

	if !sender.IsAvailable(ctx) {
		return fmt.Errorf("sender for channel %v is not available", notification.Channel)
	}

	// Schedule retry
	nextRetryAt := time.Now().Add(uc.RetryDelay)
	notification.ScheduleRetry(nextRetryAt)

	// Save notification with retry scheduled
	_, err = uc.NotificationWriter.Save(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to schedule retry: %w", err)
	}

	// Attempt to send
	if err := sender.Send(ctx, notification); err != nil {
		slog.ErrorContext(ctx, "retry failed", "error", err, "notification_id", notificationID)
		notification.MarkAsFailed(err.Error())
		_, saveErr := uc.NotificationWriter.Save(ctx, notification)
		if saveErr != nil {
			slog.ErrorContext(ctx, "failed to save failed notification", "error", saveErr)
		}
		return fmt.Errorf("retry failed: %w", err)
	}

	// Update notification status
	_, err = uc.NotificationWriter.Save(ctx, notification)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification status", "error", err)
	}

	slog.InfoContext(ctx, "notification retried successfully",
		"notification_id", notificationID,
		"retry_count", notification.RetryCount)

	return nil
}

// RetryAllFailedNotifications retries all failed notifications that are eligible for retry
func (uc *RetryFailedNotificationUseCase) RetryAllFailedNotifications(ctx context.Context) (int, []error) {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return 0, []error{fmt.Errorf("only administrators can retry notifications")}
	}

	// Get all failed notifications
	failedNotifications, err := uc.NotificationReader.FindFailedNotifications(ctx)
	if err != nil {
		return 0, []error{fmt.Errorf("failed to get failed notifications: %w", err)}
	}

	var retriedCount int
	var errors []error

	for _, notification := range failedNotifications {
		if !notification.CanRetry() {
			continue
		}

		// Check if it's time to retry
		if notification.NextRetryAt != nil && time.Now().Before(*notification.NextRetryAt) {
			continue
		}

		if err := uc.Execute(ctx, notification.ID); err != nil {
			errors = append(errors, fmt.Errorf("failed to retry notification %v: %w", notification.ID, err))
		} else {
			retriedCount++
		}
	}

	slog.InfoContext(ctx, "retried failed notifications",
		"total_failed", len(failedNotifications),
		"retried", retriedCount,
		"errors", len(errors))

	return retriedCount, errors
}
