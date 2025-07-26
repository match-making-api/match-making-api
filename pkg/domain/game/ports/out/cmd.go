package out

import (
	"context"

	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
)

// GameWriter interface for writing game data
type GameWriter interface {
	Create(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error)
	Update(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error)
	Put(ctx context.Context, gameID uuid.UUID, game *game_entities.Game) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GameModeWriter interface for writing game mode data
type GameModeWriter interface {
	Create(ctx context.Context, game *game_entities.GameMode) (*game_entities.GameMode, error)
	Update(ctx context.Context, game *game_entities.GameMode) (*game_entities.GameMode, error)
	Put(ctx context.Context, gameID uuid.UUID, game *game_entities.GameMode) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// RegionWriter interface for writing region data
type RegionWriter interface {
	Create(ctx context.Context, game *game_entities.Region) (*game_entities.Region, error)
	Update(ctx context.Context, game *game_entities.Region) (*game_entities.Region, error)
	Put(ctx context.Context, gameID uuid.UUID, game *game_entities.Region) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
