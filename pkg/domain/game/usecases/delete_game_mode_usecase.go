package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type DeleteGameModeUseCase struct {
	GameModeWriter out.GameModeWriter
	GameModeReader out.GameModeReader
}

func NewDeleteGameModeUseCase(gameModeWriter out.GameModeWriter, gameModeReader out.GameModeReader) in.DeleteGameModeCommand {
	return &DeleteGameModeUseCase{
		GameModeWriter: gameModeWriter,
		GameModeReader: gameModeReader,
	}
}

func InjectDeleteGameMode(c container.Container) error {
	c.Singleton(func(gameModeWriter out.GameModeWriter, gameModeReader out.GameModeReader) (in.DeleteGameModeCommand, error) {
		return NewDeleteGameModeUseCase(gameModeWriter, gameModeReader), nil
	})
	return nil
}

func (usecase *DeleteGameModeUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	// Check if the game mode exists
	existingGameMode, err := usecase.GameModeReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "game mode not found for deletion", "game_mode_id", id, "error", err)
		return fmt.Errorf("game mode not found: %w", err)
	}

	// Audit log before deletion
	resourceOwner := common.GetResourceOwner(ctx)
	slog.InfoContext(ctx, "deleting game mode", "game_mode_id", id, "name", existingGameMode.Name, "user_id", resourceOwner.UserID)

	// Delete the game mode
	err = usecase.GameModeWriter.Delete(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete game mode", "error", err, "game_mode_id", id)
		return fmt.Errorf("failed to delete game mode: %w", err)
	}

	slog.InfoContext(ctx, "game mode deleted successfully", "game_mode_id", id, "name", existingGameMode.Name)

	return nil
}
