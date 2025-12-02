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

type DeleteGameUseCase struct {
	GameWriter out.GameWriter
	GameReader out.GameReader
}

func NewDeleteGameUseCase(gameWriter out.GameWriter, gameReader out.GameReader) in.DeleteGameCommand {
	return &DeleteGameUseCase{
		GameWriter: gameWriter,
		GameReader: gameReader,
	}
}

func InjectDeleteGame(c container.Container) error {
	c.Singleton(func(gameWriter out.GameWriter, gameReader out.GameReader) (in.DeleteGameCommand, error) {
		return NewDeleteGameUseCase(gameWriter, gameReader), nil
	})
	return nil
}

func (usecase *DeleteGameUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	// Check if the game exists
	existingGame, err := usecase.GameReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "game not found for deletion", "game_id", id, "error", err)
		return fmt.Errorf("game not found: %w", err)
	}

	// Audit log before deletion
	resourceOwner := common.GetResourceOwner(ctx)
	slog.InfoContext(ctx, "deleting game", "game_id", id, "name", existingGame.Name, "user_id", resourceOwner.UserID)

	// Check if the game is enabled (soft delete - disable instead of delete)
	// To avoid disrupting ongoing matchmaking sessions, we just disable it
	if existingGame.Enabled {
		slog.WarnContext(ctx, "disabling game instead of deleting to avoid disrupting matchmaking", "game_id", id)
		existingGame.Enabled = false
		_, err = usecase.GameWriter.Update(ctx, existingGame)
		if err != nil {
			slog.ErrorContext(ctx, "failed to disable game", "error", err, "game_id", id)
			return fmt.Errorf("failed to disable game: %w", err)
		}
		slog.InfoContext(ctx, "game disabled successfully", "game_id", id, "name", existingGame.Name)
		return nil
	}

	// If already disabled, then delete permanently
	err = usecase.GameWriter.Delete(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete game", "error", err, "game_id", id)
		return fmt.Errorf("failed to delete game: %w", err)
	}

	slog.InfoContext(ctx, "game deleted successfully", "game_id", id, "name", existingGame.Name)

	return nil
}
