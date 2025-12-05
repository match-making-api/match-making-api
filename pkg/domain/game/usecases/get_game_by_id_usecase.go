package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golobby/container/v3"
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

func InjectGetGameByID(c container.Container) error {
	c.Singleton(func(gameReader out.GameReader) (in.GetGameByIDQuery, error) {
		return NewGetGameByIDUseCase(gameReader), nil
	})
	return nil
}

func (usecase *GetGameByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (*game_entities.Game, error) {
	game, err := usecase.GameReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get game by id", "game_id", id, "error", err)
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	slog.InfoContext(ctx, "game retrieved successfully", "game_id", id, "name", game.Name)

	return game, nil
}
