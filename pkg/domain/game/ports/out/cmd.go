package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
)

// GameWriter interface for writing game data
type GameWriter interface {
	Create(ctx context.Context, game *entities.Game) (*entities.Game, error)
	Update(ctx context.Context, game *entities.Game) (*entities.Game, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GameReader interface for reading game data
type GameReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Game, error)
	Search(ctx context.Context, query interface{}) ([]*entities.Game, error)
}

// GameModeWriter interface for writing game mode data
type GameModeWriter interface {
	Create(ctx context.Context, gameMode *entities.GameMode) (*entities.GameMode, error)
	Update(ctx context.Context, gameMode *entities.GameMode) (*entities.GameMode, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GameModeReader interface for reading game mode data
type GameModeReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.GameMode, error)
	Search(ctx context.Context, query interface{}) ([]*entities.GameMode, error)
}

// RegionWriter interface for writing region data
type RegionWriter interface {
	Create(ctx context.Context, region *entities.Region) (*entities.Region, error)
	Update(ctx context.Context, region *entities.Region) (*entities.Region, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// RegionReader interface for reading region data
type RegionReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Region, error)
	Search(ctx context.Context, query interface{}) ([]*entities.Region, error)
}
