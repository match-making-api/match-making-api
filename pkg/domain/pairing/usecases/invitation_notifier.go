package usecases

import (
	"context"

	"github.com/google/uuid"
)

// InvitationNotifier is an interface for sending invitation notifications
type InvitationNotifier interface {
	NotifyInvitationCreated(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID, message string) error
	NotifyInvitationAccepted(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error
	NotifyInvitationDeclined(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error
	NotifyInvitationRevoked(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error
}

// NoOpInvitationNotifier is a no-operation implementation of InvitationNotifier
// This can be used as a placeholder until a real notification system is implemented
type NoOpInvitationNotifier struct{}

// NewNoOpInvitationNotifier creates a new NoOpInvitationNotifier
func NewNoOpInvitationNotifier() InvitationNotifier {
	return &NoOpInvitationNotifier{}
}

// NotifyInvitationCreated is a no-op implementation
func (n *NoOpInvitationNotifier) NotifyInvitationCreated(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID, message string) error {
	// No-op: This implementation does not send any notifications
	// Replace this with actual notification logic (email, SMS, push notification, etc.)
	return nil
}

// NotifyInvitationAccepted is a no-op implementation
func (n *NoOpInvitationNotifier) NotifyInvitationAccepted(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	return nil
}

// NotifyInvitationDeclined is a no-op implementation
func (n *NoOpInvitationNotifier) NotifyInvitationDeclined(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	return nil
}

// NotifyInvitationRevoked is a no-op implementation
func (n *NoOpInvitationNotifier) NotifyInvitationRevoked(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	return nil
}
