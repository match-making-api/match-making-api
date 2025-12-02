package in

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
)

// GetGameByIDQuery interface for reading game information by ID
type GetGameByIDQuery interface {
	Execute(ctx context.Context, id uuid.UUID) (*entities.Game, error)
}

// SearchGamesQuery interface for search games
type SearchGamesQuery interface {
	Execute(ctx context.Context) ([]*entities.Game, error)
}

type GetGameModesQuery interface {
	Execute(ctx context.Context, gameID uuid.UUID) ([]*entities.GameMode, error)
}

type GetRegionsQuery interface {
	Execute(ctx context.Context, gameID uuid.UUID) ([]*entities.Region, error)
}
