package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// ListExternalInvitationsUseCase handles listing external invitations
type ListExternalInvitationsUseCase struct {
	ExternalInvitationReader pairing_out.ExternalInvitationReader
}

// ListExternalInvitationsFilter represents the filter options for listing invitations
type ListExternalInvitationsFilter struct {
	Email     *string
	MatchID   *uuid.UUID
	EventID   *uuid.UUID
	CreatedBy *uuid.UUID
	Status    *pairing_entities.ExternalInvitationStatus
}

// Execute lists external invitations based on the provided filter
func (uc *ListExternalInvitationsUseCase) Execute(ctx context.Context, filter ListExternalInvitationsFilter) ([]*pairing_entities.ExternalInvitation, error) {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return nil, fmt.Errorf("only administrators can list external invitations")
	}

	var invitations []*pairing_entities.ExternalInvitation
	var err error

	if filter.Email != nil {
		invitations, err = uc.ExternalInvitationReader.FindByEmail(ctx, *filter.Email)
	} else if filter.MatchID != nil {
		invitations, err = uc.ExternalInvitationReader.FindByMatchID(ctx, *filter.MatchID)
	} else if filter.EventID != nil {
		invitations, err = uc.ExternalInvitationReader.FindByEventID(ctx, *filter.EventID)
	} else if filter.CreatedBy != nil {
		invitations, err = uc.ExternalInvitationReader.FindByCreatedBy(ctx, *filter.CreatedBy)
	} else {
		return nil, fmt.Errorf("at least one filter parameter is required")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list external invitations: %w", err)
	}

	// Filter by status if provided
	if filter.Status != nil {
		filtered := make([]*pairing_entities.ExternalInvitation, 0)
		for _, inv := range invitations {
			if inv.Status == *filter.Status {
				filtered = append(filtered, inv)
			}
		}
		invitations = filtered
	}

	return invitations, nil
}
