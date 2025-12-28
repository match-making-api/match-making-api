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

// AcceptInvitationUseCase handles accepting an invitation
type AcceptInvitationUseCase struct {
	InvitationReader pairing_out.InvitationReader
	InvitationWriter pairing_out.InvitationWriter
	Notifier         InvitationNotifier // Optional: if nil, notifications are skipped
}

// Execute accepts an invitation if it's valid and not expired
func (uc *AcceptInvitationUseCase) Execute(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	// Get the invitation
	invitation, err := uc.InvitationReader.GetByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("failed to get invitation %v: %w", invitationID, err)
	}

	// Verify that the invitation belongs to the user
	if invitation.UserID != userID {
		return fmt.Errorf("invitation %v does not belong to user %v", invitationID, userID)
	}

	// Check if invitation can be accepted
	if !invitation.CanAccept() {
		if invitation.IsExpired() {
			// Update status to expired if not already updated
			if invitation.Status != pairing_entities.InvitationStatusExpired {
				invitation.Status = pairing_entities.InvitationStatusExpired
				_, err = uc.InvitationWriter.Save(ctx, invitation)
				if err != nil {
					slog.ErrorContext(ctx, "failed to update expired invitation", "error", err, "invitation_id", invitationID)
				}
			}
			return fmt.Errorf("invitation %v has expired", invitationID)
		}
		return fmt.Errorf("invitation %v cannot be accepted (current status: %v)", invitationID, invitation.Status)
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = pairing_entities.InvitationStatusAccepted
	invitation.AcceptedAt = &now

	_, err = uc.InvitationWriter.Save(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save accepted invitation", "error", err, "invitation_id", invitationID)
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	slog.InfoContext(ctx, "invitation accepted successfully",
		"invitation_id", invitationID,
		"user_id", userID,
		"match_id", invitation.MatchID,
		"event_id", invitation.EventID)

	// Send notification about acceptance
	if uc.Notifier != nil {
		if err := uc.Notifier.NotifyInvitationAccepted(ctx, invitationID, userID); err != nil {
			slog.ErrorContext(ctx, "failed to send acceptance notification",
				"error", err, "invitation_id", invitationID, "user_id", userID)
			// Don't fail acceptance if notification fails
		}
	}

	return nil
}
