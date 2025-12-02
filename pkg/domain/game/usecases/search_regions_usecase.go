package usecases

import (
	"context"

	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type SearchRegionsUseCase struct {
	RegionReader out.RegionReader
}

func NewSearchRegionsUseCase(regionReader out.RegionReader) in.SearchRegionsQuery {
	return &SearchRegionsUseCase{
		RegionReader: regionReader,
	}
}

func InjectSearchRegions(c container.Container) error {
	c.Singleton(func(regionReader out.RegionReader) (in.SearchRegionsQuery, error) {
		return NewSearchRegionsUseCase(regionReader), nil
	})
	return nil
}

func (usecase *SearchRegionsUseCase) Execute(ctx context.Context) ([]*game_entities.Region, error) {
	return usecase.RegionReader.Search(ctx, common.Search{})
}
