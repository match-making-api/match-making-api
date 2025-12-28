package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// DeclineInvitationUseCase handles declining an invitation
type DeclineInvitationUseCase struct {
	InvitationReader pairing_out.InvitationReader
	InvitationWriter pairing_out.InvitationWriter
	Notifier         InvitationNotifier // Optional: if nil, notifications are skipped
}

// Execute declines an invitation if it's valid and not expired
func (uc *DeclineInvitationUseCase) Execute(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	// Get the invitation
	invitation, err := uc.InvitationReader.GetByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("failed to get invitation %v: %w", invitationID, err)
	}

	// Verify that the invitation belongs to the user
	if invitation.UserID != userID {
		return fmt.Errorf("invitation %v does not belong to user %v", invitationID, userID)
	}

	// Check if invitation can be declined
	if !invitation.CanDecline() {
		if invitation.IsExpired() {
			return fmt.Errorf("invitation %v has expired", invitationID)
		}
		return fmt.Errorf("invitation %v cannot be declined (current status: %v)", invitationID, invitation.Status)
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = pairing_entities.InvitationStatusDeclined
	invitation.DeclinedAt = &now

	_, err = uc.InvitationWriter.Save(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save declined invitation", "error", err, "invitation_id", invitationID)
		return fmt.Errorf("failed to decline invitation: %w", err)
	}

	slog.InfoContext(ctx, "invitation declined successfully",
		"invitation_id", invitationID,
		"user_id", userID,
		"match_id", invitation.MatchID,
		"event_id", invitation.EventID)

	// Send notification about decline
	if uc.Notifier != nil {
		if err := uc.Notifier.NotifyInvitationDeclined(ctx, invitationID, userID); err != nil {
			slog.ErrorContext(ctx, "failed to send decline notification",
				"error", err, "invitation_id", invitationID, "user_id", userID)
			// Don't fail decline if notification fails
		}
	}

	return nil
}
