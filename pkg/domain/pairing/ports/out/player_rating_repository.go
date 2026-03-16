package pairing_out

import (
	"context"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
)

// PlayerRatingRepository persists and retrieves player ratings with resource ownership.
type PlayerRatingRepository interface {
	// GetByPlayer fetches the rating for a player (game_id can be empty for global).
	// Returns nil if not found; default MMR can be used (e.g. 1000).
	GetByPlayer(ctx context.Context, playerID, gameID, tenantID, clientID string) (*entities.PlayerRating, error)

	// Save updates the player's rating (upsert by player_id+game_id+tenant_id+client_id).
	Save(ctx context.Context, rating *entities.PlayerRating) error

	// SaveAuditEntry persists a rating change for audit trail.
	SaveAuditEntry(ctx context.Context, entry *entities.RatingAuditEntry) error
}
