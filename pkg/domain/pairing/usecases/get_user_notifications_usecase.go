package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// GetUserNotificationsUseCase handles retrieving notifications for a user
type GetUserNotificationsUseCase struct {
	NotificationReader pairing_out.NotificationReader
}

// GetUserNotificationsResult contains the notifications and total count
type GetUserNotificationsResult struct {
	Notifications []*pairing_entities.Notification
	Total        int
	Limit        int
	Offset       int
}

// Execute retrieves notifications for a user
func (uc *GetUserNotificationsUseCase) Execute(ctx context.Context, userID uuid.UUID, limit int, offset int) (*GetUserNotificationsResult, error) {
	// Verify user can access their own notifications
	currentUserIDValue := ctx.Value(common.UserIDKey)
	currentUserID, ok := currentUserIDValue.(uuid.UUID)
	if !ok || (currentUserID != userID && !common.IsAdmin(ctx)) {
		return nil, fmt.Errorf("user can only access their own notifications")
	}

	// Set defaults
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	if offset < 0 {
		offset = 0
	}

	// Get notifications
	notifications, err := uc.NotificationReader.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Get total count
	total, err := uc.NotificationReader.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification count: %w", err)
	}

	return &GetUserNotificationsResult{
		Notifications: notifications,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	}, nil
}
