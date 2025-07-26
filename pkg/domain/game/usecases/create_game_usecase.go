package usecases

import (
	"context"

	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type CreateGameUseCase struct {
	GameWriter out.GameWriter
}

func NewCreateGameUseCase(gameWriter out.GameWriter) in.CreateGameCommand {
	return &CreateGameUseCase{
		GameWriter: gameWriter,
	}
}

func (usecase *CreateGameUseCase) Execute(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error) {
	// Adicione lógica de validação aqui, se necessário
	game.ID = uuid.New() // Gera um novo ID para o jogo

	createdGame, err := usecase.GameWriter.Create(ctx, game)
	if err != nil {
		return nil, err
	}

	return createdGame, nil
}
