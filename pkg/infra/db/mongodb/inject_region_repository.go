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

// regionWriterAdapter adapts RegionRepository to game_out.RegionWriter
type regionWriterAdapter struct {
	repo RegionRepository
}

func (a *regionWriterAdapter) Create(ctx context.Context, region *game_entities.Region) (*game_entities.Region, error) {
	return a.repo.Create(ctx, region)
}

func (a *regionWriterAdapter) Update(ctx context.Context, region *game_entities.Region) (*game_entities.Region, error) {
	return a.repo.Update(ctx, region)
}

func (a *regionWriterAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.repo.Delete(ctx, id)
}

// regionReaderAdapter adapts RegionRepository to game_out.RegionReader
type regionReaderAdapter struct {
	repo RegionRepository
}

func (a *regionReaderAdapter) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.Region, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *regionReaderAdapter) Search(ctx context.Context, query interface{}) ([]*game_entities.Region, error) {
	// If query is nil, return all regions
	if query == nil {
		return a.repo.Search(ctx, common.Search{})
	}
	// If it's a common.Search, use it directly
	if s, ok := query.(common.Search); ok {
		return a.repo.Search(ctx, s)
	}
	// Otherwise, return all regions
	return a.repo.Search(ctx, common.Search{})
}

// InjectRegionRepository registers RegionRepository as a singleton in the container
func InjectRegionRepository(c container.Container) error {
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) (RegionRepository, error) {
		return NewRegionRepository(client, cfg.MongoDB.DBName, "regions"), nil
	})

	if err != nil {
		slog.Error("Failed to register RegionRepository")
		return err
	}

	// Register RegionWriter interface for usecases
	err = c.Singleton(func(repo RegionRepository) (game_out.RegionWriter, error) {
		return &regionWriterAdapter{repo: repo}, nil
	})
	if err != nil {
		slog.Error("Failed to register RegionWriter")
		return err
	}

	// Register RegionReader interface for usecases
	err = c.Singleton(func(repo RegionRepository) (game_out.RegionReader, error) {
		return &regionReaderAdapter{repo: repo}, nil
	})
	if err != nil {
		slog.Error("Failed to register RegionReader")
		return err
	}

	return nil
}
