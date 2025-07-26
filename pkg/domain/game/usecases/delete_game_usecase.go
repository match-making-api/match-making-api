package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type DeleteGameUseCase struct {
	GameWriter out.GameWriter
}

func NewDeleteGameUseCase(gameWriter out.GameWriter) in.DeleteGameCommand {
	return &DeleteGameUseCase{
		GameWriter: gameWriter,
	}
}

func (usecase *DeleteGameUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return usecase.GameWriter.Delete(ctx, id)
}
