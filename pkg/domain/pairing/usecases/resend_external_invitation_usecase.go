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

// ResendExternalInvitationUseCase handles resending an external invitation email
type ResendExternalInvitationUseCase struct {
	ExternalInvitationReader pairing_out.ExternalInvitationReader
	ExternalInvitationWriter pairing_out.ExternalInvitationWriter
	Notifier                 ExternalInvitationNotifier
}

// Execute resends the invitation email for an existing external invitation
func (uc *ResendExternalInvitationUseCase) Execute(ctx context.Context, invitationID uuid.UUID) error {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return fmt.Errorf("only administrators can resend external invitations")
	}

	// Get the invitation
	invitation, err := uc.ExternalInvitationReader.GetByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("failed to get external invitation: %w", err)
	}

	// Check if invitation can be resent
	if invitation.Status != pairing_entities.ExternalInvitationStatusPending {
		return fmt.Errorf("can only resend pending invitations")
	}

	if invitation.IsExpired() {
		return fmt.Errorf("cannot resend expired invitation")
	}

	// Resend notification
	if uc.Notifier != nil {
		if err := uc.Notifier.NotifyInvitationCreated(ctx, invitation.ID, invitation.Email, invitation.FullName, invitation.Message, invitation.RegistrationToken, invitation.MatchID, invitation.EventID); err != nil {
			slog.ErrorContext(ctx, "failed to resend external invitation notification",
				"error", err, "invitation_id", invitation.ID, "email", invitation.Email)
			return fmt.Errorf("failed to resend invitation: %w", err)
		}
	}

	slog.InfoContext(ctx, "external invitation resent successfully",
		"invitation_id", invitation.ID,
		"email", invitation.Email)

	return nil
}
