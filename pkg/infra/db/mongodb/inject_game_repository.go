package mongodb

import (
	"context"
	"log/slog"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	game_out "github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// gameWriterAdapter adapts GameRepository to game_out.GameWriter
type gameWriterAdapter struct {
	repo GameRepository
}

func (a *gameWriterAdapter) Create(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error) {
	return a.repo.Create(ctx, game)
}

func (a *gameWriterAdapter) Update(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error) {
	return a.repo.Update(ctx, game)
}

func (a *gameWriterAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.repo.Delete(ctx, id)
}

// gameReaderAdapter adapts GameRepository to game_out.GameReader
type gameReaderAdapter struct {
	repo GameRepository
}

func (a *gameReaderAdapter) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.Game, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *gameReaderAdapter) Search(ctx context.Context, query interface{}) ([]*game_entities.Game, error) {
	// If query is nil, return all games
	if query == nil {
		return a.repo.Search(ctx, common.Search{})
	}
	// If it's a common.Search, use it directly
	if s, ok := query.(common.Search); ok {
		return a.repo.Search(ctx, s)
	}
	// Otherwise, return all games
	return a.repo.Search(ctx, common.Search{})
}

// InjectGameRepository registers GameRepository as a singleton in the container
func InjectGameRepository(c container.Container) error {
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) (GameRepository, error) {
		return NewGameRepository(client, cfg.MongoDB.DBName, "games"), nil
	})

	if err != nil {
		slog.Error("Failed to register GameRepository")
		return err
	}

	// Register GameWriter interface for usecases
	err = c.Singleton(func(repo GameRepository) (game_out.GameWriter, error) {
		return &gameWriterAdapter{repo: repo}, nil
	})
	if err != nil {
		slog.Error("Failed to register GameWriter")
		return err
	}

	// Register GameReader interface for usecases
	err = c.Singleton(func(repo GameRepository) (game_out.GameReader, error) {
		return &gameReaderAdapter{repo: repo}, nil
	})
	if err != nil {
		slog.Error("Failed to register GameReader")
		return err
	}

	return nil
}
