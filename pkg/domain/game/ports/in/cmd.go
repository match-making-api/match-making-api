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

// GetGameByIDQuery interface for reading game information by ID
type GetGameByIDQuery interface {
	Execute(ctx context.Context, id uuid.UUID) (*game_entities.Game, error)
}

// SearchGamesQuery interface for search games
type SearchGamesQuery interface {
	Execute(ctx context.Context) ([]*game_entities.Game, error)
}

// CreateGameModeCommand interface for creating a new game mode
type CreateGameModeCommand interface {
	Execute(ctx context.Context, game *game_entities.GameMode) (*game_entities.GameMode, error)
}

// UpdateGameCommand interface for updating an existing game mode
type UpdateGameModeCommand interface {
	Execute(ctx context.Context, id uuid.UUID, game *game_entities.Game) (*game_entities.GameMode, error)
}

// DeleteGameCommand interface for deleting a game mode
type DeleteGameModeCommand interface {
	Execute(ctx context.Context, id uuid.UUID) error
}

// GetGameModeByIDQuery interface for reading game mode information by ID
type GetGameModeByIDQuery interface {
	Execute(ctx context.Context, id uuid.UUID) (*game_entities.GameMode, error)
}

// SearchGameModesQuery interface for search games modes
type SearchGameModesQuery interface {
	Execute(ctx context.Context) ([]*game_entities.GameMode, error)
}
