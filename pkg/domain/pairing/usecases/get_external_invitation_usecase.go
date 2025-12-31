package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// GetExternalInvitationUseCase handles retrieving an external invitation by ID
type GetExternalInvitationUseCase struct {
	ExternalInvitationReader pairing_out.ExternalInvitationReader
}

// Execute retrieves an external invitation by ID
func (uc *GetExternalInvitationUseCase) Execute(ctx context.Context, invitationID uuid.UUID) (*pairing_entities.ExternalInvitation, error) {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return nil, fmt.Errorf("only administrators can view external invitations")
	}

	invitation, err := uc.ExternalInvitationReader.GetByID(ctx, invitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get external invitation: %w", err)
	}

	return invitation, nil
}

// GetExternalInvitationByTokenUseCase handles retrieving an external invitation by registration token
type GetExternalInvitationByTokenUseCase struct {
	ExternalInvitationReader pairing_out.ExternalInvitationReader
}

// Execute retrieves an external invitation by registration token
func (uc *GetExternalInvitationByTokenUseCase) Execute(ctx context.Context, token string) (*pairing_entities.ExternalInvitation, error) {
	if token == "" {
		return nil, fmt.Errorf("registration token is required")
	}

	invitation, err := uc.ExternalInvitationReader.GetByRegistrationToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get external invitation: %w", err)
	}

	return invitation, nil
}
