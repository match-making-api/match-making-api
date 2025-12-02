package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
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

func (usecase *GetRegionsUseCase) Execute(ctx context.Context, gameID uuid.UUID) ([]*entities.Region, error) {
	return usecase.RegionReader.Search(ctx, common.Search{})
}
