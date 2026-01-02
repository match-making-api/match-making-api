package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type UpdateGameModeUseCase struct {
	GameModeWriter out.GameModeWriter
	GameModeReader out.GameModeReader
}

func NewUpdateGameModeUseCase(gameModeWriter out.GameModeWriter, gameModeReader out.GameModeReader) in.UpdateGameModeCommand {
	return &UpdateGameModeUseCase{
		GameModeWriter: gameModeWriter,
		GameModeReader: gameModeReader,
	}
}

func InjectUpdateGameMode(c container.Container) error {
	c.Singleton(func(gameModeWriter out.GameModeWriter, gameModeReader out.GameModeReader) (in.UpdateGameModeCommand, error) {
		return NewUpdateGameModeUseCase(gameModeWriter, gameModeReader), nil
	})
	return nil
}

func (usecase *UpdateGameModeUseCase) Execute(ctx context.Context, id uuid.UUID, gameMode *game_entities.GameMode) (*game_entities.GameMode, error) {
	// Get existing game mode
	existingGameMode, err := usecase.GameModeReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "game mode not found", "game_mode_id", id, "error", err)
		return nil, fmt.Errorf("game mode not found: %w", err)
	}

	// Validate game mode data
	if err := validateGameMode(gameMode); err != nil {
		slog.ErrorContext(ctx, "game mode validation failed", "error", err, "game_mode_id", id)
		return nil, fmt.Errorf("invalid game mode data: %w", err)
	}

	// Check if another game mode with the same name already exists for this game (except the current one)
	existingGameModes, err := usecase.GameModeReader.Search(ctx, nil)
	if err == nil {
		for _, existing := range existingGameModes {
			if existing.ID != id && existing.GameID == gameMode.GameID && existing.Name == gameMode.Name {
				slog.WarnContext(ctx, "game mode with same name already exists for game", "name", gameMode.Name, "game_id", gameMode.GameID, "existing_id", existing.ID)
				return nil, fmt.Errorf("game mode with name '%s' already exists for this game", gameMode.Name)
			}
		}
	}

	// Audit log before update
	resourceOwner := common.GetResourceOwner(ctx)
	slog.InfoContext(ctx, "updating game mode", "game_mode_id", id, "name", gameMode.Name, "user_id", resourceOwner.UserID)

	// Update all fields
	existingGameMode.Name = gameMode.Name
	existingGameMode.Description = gameMode.Description
	existingGameMode.GameID = gameMode.GameID
	existingGameMode.UpdatedAt = time.Now()

	// Update the game mode
	updatedGameMode, err := usecase.GameModeWriter.Update(ctx, existingGameMode)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update game mode", "error", err, "game_mode_id", id)
		return nil, fmt.Errorf("failed to update game mode: %w", err)
	}

	slog.InfoContext(ctx, "game mode updated successfully", "game_mode_id", updatedGameMode.ID, "name", updatedGameMode.Name)

	return updatedGameMode, nil
}
