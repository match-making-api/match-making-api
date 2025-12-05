package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type GetRegionByIDUseCase struct {
	RegionReader out.RegionReader
}

func NewGetRegionByIDUseCase(regionReader out.RegionReader) in.GetRegionByIDQuery {
	return &GetRegionByIDUseCase{
		RegionReader: regionReader,
	}
}

func InjectGetRegionByID(c container.Container) error {
	c.Singleton(func(regionReader out.RegionReader) (in.GetRegionByIDQuery, error) {
		return NewGetRegionByIDUseCase(regionReader), nil
	})
	return nil
}

func (usecase *GetRegionByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (*game_entities.Region, error) {
	region, err := usecase.RegionReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get region by id", "region_id", id, "error", err)
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	slog.InfoContext(ctx, "region retrieved successfully", "region_id", id, "name", region.Name)

	return region, nil
}
