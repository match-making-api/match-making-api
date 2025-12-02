package usecases

import (
	"context"

	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type SearchGameModesUseCase struct {
	GameModeReader out.GameModeReader
}

func NewSearchGameModesUseCase(gameModeReader out.GameModeReader) in.SearchGameModesQuery {
	return &SearchGameModesUseCase{
		GameModeReader: gameModeReader,
	}
}

func InjectSearchGameModes(c container.Container) error {
	c.Singleton(func(gameModeReader out.GameModeReader) (in.SearchGameModesQuery, error) {
		return NewSearchGameModesUseCase(gameModeReader), nil
	})
	return nil
}

func (usecase *SearchGameModesUseCase) Execute(ctx context.Context) ([]*game_entities.GameMode, error) {
	return usecase.GameModeReader.Search(ctx, common.Search{})
}
