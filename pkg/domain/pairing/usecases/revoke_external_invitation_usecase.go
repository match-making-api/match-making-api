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

// RevokeExternalInvitationUseCase handles revoking an external invitation
type RevokeExternalInvitationUseCase struct {
	ExternalInvitationReader pairing_out.ExternalInvitationReader
	ExternalInvitationWriter pairing_out.ExternalInvitationWriter
	Notifier                 ExternalInvitationNotifier
}

// Execute revokes an external invitation
func (uc *RevokeExternalInvitationUseCase) Execute(ctx context.Context, invitationID uuid.UUID) error {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return fmt.Errorf("only administrators can revoke external invitations")
	}

	// Get the invitation
	invitation, err := uc.ExternalInvitationReader.GetByID(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("failed to get external invitation: %w", err)
	}

	// Check if invitation can be revoked
	if !invitation.CanRevoke() {
		return fmt.Errorf("invitation cannot be revoked (status: %v)", invitation.Status)
	}

	// Revoke the invitation
	resourceOwner := common.GetResourceOwner(ctx)
	revokedBy := resourceOwner.UserID
	now := time.Now()

	invitation.Status = pairing_entities.ExternalInvitationStatusRevoked
	invitation.RevokedAt = &now
	invitation.RevokedBy = &revokedBy

	// Save the updated invitation
	_, err = uc.ExternalInvitationWriter.Save(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save revoked external invitation", "error", err, "invitation_id", invitationID)
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	slog.InfoContext(ctx, "external invitation revoked successfully",
		"invitation_id", invitation.ID,
		"email", invitation.Email,
		"revoked_by", revokedBy)

	return nil
}
