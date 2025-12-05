package usecases

import (
	"context"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type GetRegionsUseCase struct {
	RegionReader out.RegionReader
}

func NewGetRegionsUseCase(regionReader out.RegionReader) in.GetRegionsQuery {
	return &GetRegionsUseCase{
		RegionReader: regionReader,
	}
}

func InjectGetRegions(c container.Container) error {
	c.Singleton(func(regionReader out.RegionReader) (in.GetRegionsQuery, error) {
		return NewGetRegionsUseCase(regionReader), nil
	})
	return nil
}

func (usecase *GetRegionsUseCase) Execute(ctx context.Context, gameID uuid.UUID) ([]*game_entities.Region, error) {
	return usecase.RegionReader.Search(ctx, common.Search{})
}
