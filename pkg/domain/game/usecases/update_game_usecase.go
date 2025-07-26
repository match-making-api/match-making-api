package usecases

import (
	"context"

	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type UpdateGameUseCase struct {
	GameWriter out.GameWriter
	GameReader out.GameReader
}

func NewUpdateGameUseCase(gameWriter out.GameWriter, gameReader out.GameReader) in.UpdateGameCommand {
	return &UpdateGameUseCase{
		GameWriter: gameWriter,
		GameReader: gameReader,
	}
}

func (usecase *UpdateGameUseCase) Execute(ctx context.Context, id uuid.UUID, game *game_entities.Game) (*game_entities.Game, error) {
	existingGame, err := usecase.GameReader.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Atualize apenas os campos necessários
	existingGame.Name = game.Name
	// Atualize outros campos conforme necessário

	updatedGame, err := usecase.GameWriter.Update(ctx, existingGame)
	if err != nil {
		return nil, err
	}

	return updatedGame, nil
}
