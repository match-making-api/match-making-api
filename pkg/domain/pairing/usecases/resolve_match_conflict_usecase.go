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

// ResolveMatchConflictUseCase allows administrators to resolve or override flagged conflicts
type ResolveMatchConflictUseCase struct {
	PairReader pairing_out.PairReader
	PairWriter pairing_out.PairWriter
}

// ResolveConflictPayload contains the information needed to resolve a conflict
type ResolveConflictPayload struct {
	PairID  uuid.UUID
	Action  ConflictResolutionAction
	Reason  string // Optional reason for the resolution
}

type ConflictResolutionAction string

const (
	ConflictResolutionResolve ConflictResolutionAction = "resolve" // Mark as resolved (conflict was fixed)
	ConflictResolutionOverride ConflictResolutionAction = "override" // Override the conflict (admin decision)
	ConflictResolutionRemove   ConflictResolutionAction = "remove"   // Remove the conflicting match
)

// Execute resolves a flagged conflict according to the specified action
func (uc *ResolveMatchConflictUseCase) Execute(ctx context.Context, payload ResolveConflictPayload) error {
	// Verify that the user is an administrator
	if !common.IsAdmin(ctx) {
		return fmt.Errorf("only administrators can resolve conflicts")
	}

	// Get the pair
	pair, err := uc.PairReader.GetByID(ctx, payload.PairID)
	if err != nil {
		return fmt.Errorf("failed to get pair %v: %w", payload.PairID, err)
	}

	// Verify the pair is actually flagged
	if pair.ConflictStatus != pairing_entities.ConflictStatusFlagged {
		return fmt.Errorf("pair %v is not flagged as conflicting", payload.PairID)
	}

	resourceOwner := common.GetResourceOwner(ctx)

	switch payload.Action {
	case ConflictResolutionResolve:
		// Mark conflict as resolved
		pair.ConflictStatus = pairing_entities.ConflictStatusResolved
		pair.ConflictReason = ""
		slog.InfoContext(ctx, "conflict resolved by admin", 
			"pair_id", payload.PairID, "admin_user_id", resourceOwner.UserID, "reason", payload.Reason)

	case ConflictResolutionOverride:
		// Override the conflict (keep the match despite the conflict)
		pair.ConflictStatus = pairing_entities.ConflictStatusNone
		pair.ConflictReason = ""
		slog.InfoContext(ctx, "conflict overridden by admin", 
			"pair_id", payload.PairID, "admin_user_id", resourceOwner.UserID, "reason", payload.Reason)

	case ConflictResolutionRemove:
		// Note: This would require a Delete method on PairWriter
		// For now, we'll just flag it differently or leave it for future implementation
		// You might want to add a PairDeleter interface or extend PairWriter
		slog.WarnContext(ctx, "conflict removal requested but not implemented", 
			"pair_id", payload.PairID, "admin_user_id", resourceOwner.UserID)
		return fmt.Errorf("conflict removal is not yet implemented")

	default:
		return fmt.Errorf("invalid conflict resolution action: %v", payload.Action)
	}

	// Save the updated pair
	_, err = uc.PairWriter.Save(pair)
	if err != nil {
		return fmt.Errorf("failed to save resolved pair %v: %w", payload.PairID, err)
	}

	slog.InfoContext(ctx, "conflict resolution completed", 
		"pair_id", payload.PairID, "action", payload.Action, "admin_user_id", resourceOwner.UserID)

	return nil
}
