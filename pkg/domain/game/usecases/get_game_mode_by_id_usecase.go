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

type GetGameModeByIDUseCase struct {
	GameModeReader out.GameModeReader
}

func NewGetGameModeByIDUseCase(gameModeReader out.GameModeReader) in.GetGameModeByIDQuery {
	return &GetGameModeByIDUseCase{
		GameModeReader: gameModeReader,
	}
}

func InjectGetGameModeByID(c container.Container) error {
	c.Singleton(func(gameModeReader out.GameModeReader) (in.GetGameModeByIDQuery, error) {
		return NewGetGameModeByIDUseCase(gameModeReader), nil
	})
	return nil
}

func (usecase *GetGameModeByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (*game_entities.GameMode, error) {
	gameMode, err := usecase.GameModeReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get game mode by id", "game_mode_id", id, "error", err)
		return nil, fmt.Errorf("failed to get game mode: %w", err)
	}

	slog.InfoContext(ctx, "game mode retrieved successfully", "game_mode_id", id, "name", gameMode.Name)

	return gameMode, nil
}
