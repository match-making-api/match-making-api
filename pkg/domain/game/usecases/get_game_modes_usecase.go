package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type GetGameModesUseCase struct {
	GameModeReader out.GameModeReader
}

func NewGetGameModesUseCase(gameModeReader out.GameModeReader) in.GetGameModesQuery {
	return &GetGameModesUseCase{
		GameModeReader: gameModeReader,
	}
}

func (uc *GetGameModesUseCase) Execute(c context.Context, gameID uuid.UUID) ([]*entities.GameMode, error) {
	return uc.GameModeReader.Search(c, entities.NewSearchGameModeByGameID(c, gameID))

	if err != nil {
		return nil, err
	}
}
