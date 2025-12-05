package in

import (
	"context"

	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
)

// CreateGameCommand interface for creating a new game
type CreateGameCommand interface {
	Execute(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error)
}

// UpdateGameCommand interface for updating an existing game
type UpdateGameCommand interface {
	Execute(ctx context.Context, id uuid.UUID, game *game_entities.Game) (*game_entities.Game, error)
}

// DeleteGameCommand interface for deleting a game
type DeleteGameCommand interface {
	Execute(ctx context.Context, id uuid.UUID) error
}

// CreateGameModeCommand interface for creating a new game mode
type CreateGameModeCommand interface {
	Execute(ctx context.Context, gameMode *game_entities.GameMode) (*game_entities.GameMode, error)
}

// UpdateGameModeCommand interface for updating an existing game mode
type UpdateGameModeCommand interface {
	Execute(ctx context.Context, id uuid.UUID, gameMode *game_entities.GameMode) (*game_entities.GameMode, error)
}

// DeleteGameModeCommand interface for deleting a game mode
type DeleteGameModeCommand interface {
	Execute(ctx context.Context, id uuid.UUID) error
}

// GetGameModeByIDQuery interface for reading game mode information by ID
type GetGameModeByIDQuery interface {
	Execute(ctx context.Context, id uuid.UUID) (*game_entities.GameMode, error)
}

// SearchGameModesQuery interface for searching game modes
type SearchGameModesQuery interface {
	Execute(ctx context.Context) ([]*game_entities.GameMode, error)
}

// CreateRegionCommand interface for creating a new region
type CreateRegionCommand interface {
	Execute(ctx context.Context, region *game_entities.Region) (*game_entities.Region, error)
}

// UpdateRegionCommand interface for updating an existing region
type UpdateRegionCommand interface {
	Execute(ctx context.Context, id uuid.UUID, region *game_entities.Region) (*game_entities.Region, error)
}

// DeleteRegionCommand interface for deleting a region
type DeleteRegionCommand interface {
	Execute(ctx context.Context, id uuid.UUID) error
}

// GetRegionByIDQuery interface for reading region information by ID
type GetRegionByIDQuery interface {
	Execute(ctx context.Context, id uuid.UUID) (*game_entities.Region, error)
}

// SearchRegionsQuery interface for searching regions
type SearchRegionsQuery interface {
	Execute(ctx context.Context) ([]*game_entities.Region, error)
}
