package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// MarkNotificationReadUseCase handles marking a notification as read
type MarkNotificationReadUseCase struct {
	NotificationReader pairing_out.NotificationReader
	NotificationWriter  pairing_out.NotificationWriter
}

// Execute marks a notification as read
func (uc *MarkNotificationReadUseCase) Execute(ctx context.Context, notificationID uuid.UUID) error {
	// Get notification
	notification, err := uc.NotificationReader.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	// Verify user can mark their own notification as read
	currentUserIDValue := ctx.Value(common.UserIDKey)
	currentUserID, ok := currentUserIDValue.(uuid.UUID)
	if !ok || (currentUserID != notification.UserID && !common.IsAdmin(ctx)) {
		return fmt.Errorf("user can only mark their own notifications as read")
	}

	// Only in-app notifications can be marked as read
	if notification.Channel != pairing_entities.NotificationChannelInApp {
		return fmt.Errorf("only in-app notifications can be marked as read")
	}

	// Mark as read
	notification.MarkAsRead()

	// Save notification
	_, err = uc.NotificationWriter.Save(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return nil
}
