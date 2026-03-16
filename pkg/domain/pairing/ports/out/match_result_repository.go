package pairing_out

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
)

// MatchResultRepository persists match results with resource ownership.
type MatchResultRepository interface {
	// Save persists a match result. Idempotent: if match_id already exists, returns existing (no duplicate).
	Save(ctx context.Context, result *entities.MatchResult) (*entities.MatchResult, error)

	// GetByMatchID returns the result for a match, or nil if not found.
	GetByMatchID(ctx context.Context, matchID uuid.UUID) (*entities.MatchResult, error)
}
