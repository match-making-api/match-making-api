package usecases

import (
	"context"

	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type SearchGamesUseCase struct {
	GameReader out.GameReader
}

func NewSearchGamesUseCase(gameReader out.GameReader) in.SearchGamesQuery {
	return &SearchGamesUseCase{
		GameReader: gameReader,
	}
}

func (usecase *SearchGamesUseCase) Execute(ctx context.Context) ([]*game_entities.Game, error) {
	return usecase.GameReader.Search(ctx, common.Search{})
}
