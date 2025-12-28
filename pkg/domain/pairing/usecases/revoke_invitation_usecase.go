package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// RevokeInvitationUseCase handles revoking an invitation (admin only)
type RevokeInvitationUseCase struct {
	InvitationReader pairing_out.InvitationReader
	InvitationWriter pairing_out.InvitationWriter
	Notifier         InvitationNotifier // Optional: if nil, notifications are skipped
}

// Execute revokes an invitation if it's still pending
func (uc *RevokeInvitationUseCase) Execute(ctx context.Context, invitationID uuid.UUID) error {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return fmt.Errorf("only administrators can revoke invitations")
	}

	resourceOwner := common.GetResourceOwner(ctx)
	revokedBy := resourceOwner.UserID

	// Get the invitation
	invitation, err := uc.InvitationReader.GetByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("failed to get invitation %v: %w", invitationID, err)
	}

	// Check if invitation can be revoked
	if !invitation.CanRevoke() {
		return fmt.Errorf("invitation %v cannot be revoked (current status: %v)", invitationID, invitation.Status)
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = pairing_entities.InvitationStatusRevoked
	invitation.RevokedAt = &now
	invitation.RevokedBy = &revokedBy

	_, err = uc.InvitationWriter.Save(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save revoked invitation", "error", err, "invitation_id", invitationID)
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	slog.InfoContext(ctx, "invitation revoked successfully",
		"invitation_id", invitationID,
		"revoked_by", revokedBy,
		"user_id", invitation.UserID)

	// Send notification to the user about revocation
	if uc.Notifier != nil {
		if err := uc.Notifier.NotifyInvitationRevoked(ctx, invitationID, invitation.UserID); err != nil {
			slog.ErrorContext(ctx, "failed to send revocation notification",
				"error", err, "invitation_id", invitationID, "user_id", invitation.UserID)
			// Don't fail revocation if notification fails
		}
	}

	return nil
}
