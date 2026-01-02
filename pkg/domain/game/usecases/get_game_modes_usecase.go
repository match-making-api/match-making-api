package usecases

import (
	"context"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
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

func InjectGetGameModes(c container.Container) error {
	c.Singleton(func(gameModeReader out.GameModeReader) (in.GetGameModesQuery, error) {
		return NewGetGameModesUseCase(gameModeReader), nil
	})
	return nil
}

func (uc *GetGameModesUseCase) Execute(c context.Context, gameID uuid.UUID) ([]*game_entities.GameMode, error) {
	return uc.GameModeReader.Search(c, game_entities.NewSearchGameModeByGameID(c, gameID.String()))
}
