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
	parties_out "github.com/leet-gaming/match-making-api/pkg/domain/parties/ports/out"
)

// CreateManualInvitationUseCase handles the creation of manual invitations by administrators
type CreateManualInvitationUseCase struct {
	InvitationWriter pairing_out.InvitationWriter
	PeerReader       parties_out.PeerReader
	PairReader       pairing_out.PairReader
	Notifier         InvitationNotifier // Optional: if nil, notifications are skipped
}

// CreateInvitationPayload contains the information needed to create an invitation
type CreateInvitationPayload struct {
	Type           pairing_entities.InvitationType
	UserID         uuid.UUID
	MatchID        *uuid.UUID
	EventID        *uuid.UUID
	Message        string
	ExpirationDate *time.Time
}

// Execute creates a manual invitation after validating all requirements
func (uc *CreateManualInvitationUseCase) Execute(ctx context.Context, payload CreateInvitationPayload) (*pairing_entities.Invitation, error) {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return nil, fmt.Errorf("only administrators can create manual invitations")
	}

	resourceOwner := common.GetResourceOwner(ctx)
	createdBy := resourceOwner.UserID

	// Validate user exists and is eligible
	if err := uc.validateUser(ctx, payload.UserID); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Validate match or event
	if err := uc.validateMatchOrEvent(ctx, payload); err != nil {
		return nil, fmt.Errorf("match/event validation failed: %w", err)
	}

	// Validate expiration date if provided
	if payload.ExpirationDate != nil && payload.ExpirationDate.Before(time.Now()) {
		return nil, fmt.Errorf("expiration date must be in the future")
	}

	// Create invitation
	invitation := pairing_entities.NewInvitation(
		resourceOwner,
		payload.Type,
		payload.UserID,
		payload.MatchID,
		payload.EventID,
		payload.Message,
		payload.ExpirationDate,
		createdBy,
	)

	// Save invitation
	savedInvitation, err := uc.InvitationWriter.Save(ctx, invitation)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save invitation", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	slog.InfoContext(ctx, "manual invitation created successfully",
		"invitation_id", savedInvitation.ID,
		"user_id", payload.UserID,
		"match_id", payload.MatchID,
		"event_id", payload.EventID,
		"created_by", createdBy,
		"expiration_date", payload.ExpirationDate)

	// Send notification to the user
	if uc.Notifier != nil {
		if err := uc.Notifier.NotifyInvitationCreated(ctx, savedInvitation.ID, payload.UserID, payload.Message); err != nil {
			slog.ErrorContext(ctx, "failed to send invitation notification",
				"error", err, "invitation_id", savedInvitation.ID, "user_id", payload.UserID)
			// Don't fail invitation creation if notification fails
		}
	}

	return savedInvitation, nil
}

// validateUser checks if the user exists and is eligible for invitations
func (uc *CreateManualInvitationUseCase) validateUser(ctx context.Context, userID uuid.UUID) error {
	// Check if peer exists (users are represented as peers in the system)
	_, err := uc.PeerReader.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user %v does not exist or is not eligible: %w", userID, err)
	}

	// Additional eligibility checks can be added here
	// For example: check if user is active, not banned, etc.

	return nil
}

// validateMatchOrEvent validates that the match or event is valid and open for new participants
func (uc *CreateManualInvitationUseCase) validateMatchOrEvent(ctx context.Context, payload CreateInvitationPayload) error {
	if payload.Type == pairing_entities.InvitationTypeMatch {
		if payload.MatchID == nil {
			return fmt.Errorf("match_id is required for match invitations")
		}

		// Validate that the match exists
		pair, err := uc.PairReader.GetByID(ctx, *payload.MatchID)
		if err != nil {
			return fmt.Errorf("match %v does not exist: %w", *payload.MatchID, err)
		}

		// Check if match is open for new participants
		// This is a simplified check - you may need to add more logic based on your business rules
		// For example: check if match has available slots, hasn't started yet, etc.
		if pair.ConflictStatus == pairing_entities.ConflictStatusFlagged {
			return fmt.Errorf("match %v has conflicts and is not open for new participants", *payload.MatchID)
		}

		// Additional validations can be added here
		// For example: check if match is full, has started, etc.
	} else if payload.Type == pairing_entities.InvitationTypeEvent {
		if payload.EventID == nil {
			return fmt.Errorf("event_id is required for event invitations")
		}

		// Event validation would go here
		// For now, we'll just check that event_id is provided
		// TODO: Implement event validation when event system is available
	} else {
		return fmt.Errorf("invalid invitation type: %v", payload.Type)
	}

	return nil
}
