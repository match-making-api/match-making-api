package usecases

import (
	"context"

	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type GetGameByIDUseCase struct {
	GameReader out.GameReader
}

func NewGetGameByIDUseCase(gameReader out.GameReader) in.GetGameByIDQuery {
	return &GetGameByIDUseCase{
		GameReader: gameReader,
	}
}

func (usecase *GetGameByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (*game_entities.Game, error) {
	return usecase.GameReader.GetByID(ctx, id)
}
