package usecases

import (
	"context"

	"github.com/google/uuid"
)

// ExternalInvitationNotifier defines the interface for sending notifications related to external invitations
type ExternalInvitationNotifier interface {
	// NotifyInvitationCreated sends an invitation email to the external user
	NotifyInvitationCreated(ctx context.Context, invitationID uuid.UUID, email string, fullName string, message string, registrationToken string, matchID *uuid.UUID, eventID *uuid.UUID) error

	// NotifyInvitationAccepted sends a notification to the administrator when an invited user completes registration
	NotifyInvitationAccepted(ctx context.Context, invitationID uuid.UUID, email string, fullName string, userID uuid.UUID) error

	// NotifyInvitationExpired sends a notification to the administrator when an invitation expires
	NotifyInvitationExpired(ctx context.Context, invitationID uuid.UUID, email string, fullName string) error
}

// NoOpExternalInvitationNotifier is a no-operation implementation of ExternalInvitationNotifier
// that does nothing when notification methods are called
type NoOpExternalInvitationNotifier struct{}

// NotifyInvitationCreated implements ExternalInvitationNotifier
func (n *NoOpExternalInvitationNotifier) NotifyInvitationCreated(ctx context.Context, invitationID uuid.UUID, email string, fullName string, message string, registrationToken string, matchID *uuid.UUID, eventID *uuid.UUID) error {
	// No-op: Replace this with actual notification logic (email, SMS, push notification, etc.)
	return nil
}

// NotifyInvitationAccepted implements ExternalInvitationNotifier
func (n *NoOpExternalInvitationNotifier) NotifyInvitationAccepted(ctx context.Context, invitationID uuid.UUID, email string, fullName string, userID uuid.UUID) error {
	// No-op: Replace this with actual notification logic (email, SMS, push notification, etc.)
	return nil
}

// NotifyInvitationExpired implements ExternalInvitationNotifier
func (n *NoOpExternalInvitationNotifier) NotifyInvitationExpired(ctx context.Context, invitationID uuid.UUID, email string, fullName string) error {
	// No-op: Replace this with actual notification logic (email, SMS, push notification, etc.)
	return nil
}
