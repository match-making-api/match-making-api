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

// gameModeWriterAdapter adapts GameModeRepository to game_out.GameModeWriter
type gameModeWriterAdapter struct {
	repo GameModeRepository
}

func (a *gameModeWriterAdapter) Create(ctx context.Context, gameMode *game_entities.GameMode) (*game_entities.GameMode, error) {
	return a.repo.Create(ctx, gameMode)
}

func (a *gameModeWriterAdapter) Update(ctx context.Context, gameMode *game_entities.GameMode) (*game_entities.GameMode, error) {
	return a.repo.Update(ctx, gameMode)
}

func (a *gameModeWriterAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.repo.Delete(ctx, id)
}

// gameModeReaderAdapter adapts GameModeRepository to game_out.GameModeReader
type gameModeReaderAdapter struct {
	repo GameModeRepository
}

func (a *gameModeReaderAdapter) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.GameMode, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *gameModeReaderAdapter) Search(ctx context.Context, query interface{}) ([]*game_entities.GameMode, error) {
	// If query is nil, return all game modes
	if query == nil {
		return a.repo.Search(ctx, common.Search{})
	}
	// If it's a common.Search, use it directly
	if s, ok := query.(common.Search); ok {
		return a.repo.Search(ctx, s)
	}
	// Otherwise, return all game modes
	return a.repo.Search(ctx, common.Search{})
}

// InjectGameModeRepository registers GameModeRepository as a singleton in the container
func InjectGameModeRepository(c container.Container) error {
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) (GameModeRepository, error) {
		return NewGameModeRepository(client, cfg.MongoDB.DBName, "game_modes"), nil
	})

	if err != nil {
		slog.Error("Failed to register GameModeRepository")
		return err
	}

	// Register GameModeWriter interface for usecases
	err = c.Singleton(func(repo GameModeRepository) (game_out.GameModeWriter, error) {
		return &gameModeWriterAdapter{repo: repo}, nil
	})
	if err != nil {
		slog.Error("Failed to register GameModeWriter")
		return err
	}

	// Register GameModeReader interface for usecases
	err = c.Singleton(func(repo GameModeRepository) (game_out.GameModeReader, error) {
		return &gameModeReaderAdapter{repo: repo}, nil
	})
	if err != nil {
		slog.Error("Failed to register GameModeReader")
		return err
	}

	return nil
}
