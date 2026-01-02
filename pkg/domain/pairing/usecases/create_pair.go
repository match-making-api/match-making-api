package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	parties_out "github.com/leet-gaming/match-making-api/pkg/domain/parties/ports/out"
)

// ConflictVerifier is an interface for verifying match conflicts
type ConflictVerifier interface {
	Execute(ctx context.Context, partyID uuid.UUID) (*ConflictResult, error)
}

type CreatePairUseCase struct {
	PartyReader      parties_out.PartyReader
	PairWriter       pairing_out.PairWriter
	ConflictVerifier ConflictVerifier // Optional: if nil, conflict checking is skipped
}

func (uc *CreatePairUseCase) Execute(ctx context.Context, partyIDs []uuid.UUID) (*pairing_entities.Pair, error) {
	resourceOwner := common.GetResourceOwner(ctx)
	pair := pairing_entities.NewPair(len(partyIDs), resourceOwner)

	for _, partyID := range partyIDs {
		party, err := uc.PartyReader.GetByID(partyID)
		if err != nil {
			return nil, fmt.Errorf("CreatePairUseCase.Execute: unable to create pair. PartyID: %v not found (Error: %v)", partyID, err)
		}
		pair.Match[partyID] = party
	}

	savedPair, err := uc.PairWriter.Save(pair)
	if err != nil {
		slog.ErrorContext(ctx, "failed to save pair", "error", err, "party_ids", partyIDs)
		return nil, fmt.Errorf("CreatePairUseCase.Execute: unable to create pair for PartyIDs %v, due to create error: %v", partyIDs, err)
	}

	slog.InfoContext(ctx, "pair created successfully", "pair_id", savedPair.ID, "party_ids", partyIDs)

	// Automatically verify conflicts for all parties in the new pair
	if uc.ConflictVerifier != nil {
		for partyID := range savedPair.Match {
			conflictResult, err := uc.ConflictVerifier.Execute(ctx, partyID)
			if err != nil {
				slog.ErrorContext(ctx, "failed to verify conflicts after pair creation", 
					"party_id", partyID, "pair_id", savedPair.ID, "error", err)
				// Don't fail pair creation if conflict check fails
				continue
			}

			if conflictResult.HasConflict {
				slog.WarnContext(ctx, "conflicts detected after pair creation", 
					"party_id", partyID, "pair_id", savedPair.ID, 
					"conflicting_pairs", conflictResult.ConflictingPairs)
			}
		}
	}

	return savedPair, nil
}
