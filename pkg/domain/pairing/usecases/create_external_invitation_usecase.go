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

// CreateExternalInvitationUseCase handles the creation of external invitations by administrators
type CreateExternalInvitationUseCase struct {
	ExternalInvitationWriter pairing_out.ExternalInvitationWriter
	ExternalInvitationReader pairing_out.ExternalInvitationReader
	PairReader               pairing_out.PairReader
	Notifier                 ExternalInvitationNotifier // Optional: if nil, notifications are skipped
}

// CreateExternalInvitationPayload contains the information needed to create an external invitation
type CreateExternalInvitationPayload struct {
	Type           pairing_entities.ExternalInvitationType
	FullName       string
	Email          string
	Message        string
	ExpirationDate *time.Time
	MatchID        *uuid.UUID
	EventID        *uuid.UUID
}

// Execute creates an external invitation after validating all requirements
func (uc *CreateExternalInvitationUseCase) Execute(ctx context.Context, payload CreateExternalInvitationPayload) (*pairing_entities.ExternalInvitation, error) {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return nil, fmt.Errorf("only administrators can create external invitations")
	}

	resourceOwner := common.GetResourceOwner(ctx)
	createdBy := resourceOwner.UserID

	// Validate email format
	if err := common.ValidateEmail(payload.Email); err != nil {
		return nil, fmt.Errorf("email validation failed: %w", err)
	}

	// Validate full name
	if payload.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	// Validate match or event
	if err := uc.validateMatchOrEvent(ctx, payload); err != nil {
		return nil, fmt.Errorf("match/event validation failed: %w", err)
	}

	// Validate expiration date if provided
	if payload.ExpirationDate != nil && payload.ExpirationDate.Before(time.Now()) {
		return nil, fmt.Errorf("expiration date must be in the future")
	}

	// Check if there's already a pending invitation for this email and match/event
	existingInvitations, err := uc.ExternalInvitationReader.FindByEmail(ctx, payload.Email)
	if err == nil {
		for _, inv := range existingInvitations {
			if inv.Status == pairing_entities.ExternalInvitationStatusPending &&
				!inv.IsExpired() &&
				((payload.MatchID != nil && inv.MatchID != nil && *inv.MatchID == *payload.MatchID) ||
					(payload.EventID != nil && inv.EventID != nil && *inv.EventID == *payload.EventID)) {
				return nil, fmt.Errorf("a pending invitation already exists for this email and match/event")
			}
		}
	}

	// Generate registration token
	registrationToken, err := common.GenerateRegistrationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate registration token: %w", err)
	}

	// Create invitation
	invitation := pairing_entities.NewExternalInvitation(
		resourceOwner,
		payload.Type,
		payload.FullName,
		payload.Email,
		payload.Message,
		payload.ExpirationDate,
		registrationToken,
		payload.MatchID,
		payload.EventID,
		createdBy,
	)

	// Save invitation
	savedInvitation, err := uc.ExternalInvitationWriter.Save(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save external invitation", "error", err, "email", payload.Email)
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	slog.InfoContext(ctx, "external invitation created successfully",
		"invitation_id", savedInvitation.ID,
		"email", payload.Email,
		"match_id", payload.MatchID,
		"event_id", payload.EventID,
		"created_by", createdBy,
		"expiration_date", payload.ExpirationDate)

	// Send invitation email
	if uc.Notifier != nil {
		if err := uc.Notifier.NotifyInvitationCreated(ctx, savedInvitation.ID, payload.Email, payload.FullName, payload.Message, registrationToken, payload.MatchID, payload.EventID); err != nil {
			slog.ErrorContext(ctx, "failed to send external invitation notification",
				"error", err, "invitation_id", savedInvitation.ID, "email", payload.Email)
			// Don't fail invitation creation if notification fails
		}
	}

	return savedInvitation, nil
}

// validateMatchOrEvent validates that the match or event is valid and open for new participants
func (uc *CreateExternalInvitationUseCase) validateMatchOrEvent(ctx context.Context, payload CreateExternalInvitationPayload) error {
	if payload.Type == pairing_entities.ExternalInvitationTypeMatch {
		if payload.MatchID == nil {
			return fmt.Errorf("match_id is required for match invitations")
		}

		// Validate that the match exists
		pair, err := uc.PairReader.GetByID(ctx, *payload.MatchID)
		if err != nil {
			return fmt.Errorf("match %v does not exist: %w", *payload.MatchID, err)
		}

		// Check if match is open for new participants
		if pair.ConflictStatus == pairing_entities.ConflictStatusFlagged {
			return fmt.Errorf("match %v has conflicts and is not open for new participants", *payload.MatchID)
		}

		// Additional validations can be added here
	} else if payload.Type == pairing_entities.ExternalInvitationTypeEvent {
		if payload.EventID == nil {
			return fmt.Errorf("event_id is required for event invitations")
		}

		// Event validation would go here
		// TODO: Implement event validation when event system is available
	} else {
		return fmt.Errorf("invalid invitation type: %v", payload.Type)
	}

	return nil
}
