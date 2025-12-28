package usecases

import (
	"context"

	"github.com/google/uuid"
)

// NoOpConflictNotifier is a no-operation implementation of ConflictNotifier
// This can be used as a placeholder until a real notification system is implemented
type NoOpConflictNotifier struct{}

// NewNoOpConflictNotifier creates a new NoOpConflictNotifier
func NewNoOpConflictNotifier() ConflictNotifier {
	return &NoOpConflictNotifier{}
}

// NotifyConflict is a no-op implementation that does nothing
func (n *NoOpConflictNotifier) NotifyConflict(ctx context.Context, partyID uuid.UUID, pairID uuid.UUID, reason string) error {
	// No-op: This implementation does not send any notifications
	// Replace this with actual notification logic (email, SMS, push notification, etc.)
	return nil
}
