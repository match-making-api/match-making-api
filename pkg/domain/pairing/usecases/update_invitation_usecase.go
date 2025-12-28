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

// UpdateInvitationUseCase handles updating an invitation (admin only, before acceptance)
type UpdateInvitationUseCase struct {
	InvitationReader pairing_out.InvitationReader
	InvitationWriter pairing_out.InvitationWriter
}

// UpdateInvitationPayload contains the fields that can be updated
type UpdateInvitationPayload struct {
	Message        *string
	ExpirationDate *time.Time
}

// Execute updates an invitation if it's still pending
func (uc *UpdateInvitationUseCase) Execute(ctx context.Context, invitationID uuid.UUID, payload UpdateInvitationPayload) (*pairing_entities.Invitation, error) {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return nil, fmt.Errorf("only administrators can update invitations")
	}

	// Get the invitation
	invitation, err := uc.InvitationReader.GetByID(ctx, invitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation %v: %w", invitationID, err)
	}

	// Only allow updates to pending invitations
	if invitation.Status != pairing_entities.InvitationStatusPending {
		return nil, fmt.Errorf("invitation %v cannot be updated (current status: %v)", invitationID, invitation.Status)
	}

	// Update fields if provided
	if payload.Message != nil {
		invitation.Message = *payload.Message
	}

	if payload.ExpirationDate != nil {
		// Validate expiration date
		if payload.ExpirationDate.Before(time.Now()) {
			return nil, fmt.Errorf("expiration date must be in the future")
		}
		invitation.ExpirationDate = payload.ExpirationDate
	}

	// Update UpdatedAt timestamp (BaseEntity handles this, but we ensure it's set)
	invitation.UpdatedAt = time.Now()

	// Save updated invitation
	savedInvitation, err := uc.InvitationWriter.Save(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save updated invitation", "error", err, "invitation_id", invitationID)
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	slog.InfoContext(ctx, "invitation updated successfully",
		"invitation_id", invitationID,
		"updated_by", common.GetResourceOwner(ctx).UserID)

	return savedInvitation, nil
}
