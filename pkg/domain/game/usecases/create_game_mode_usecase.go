package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid"
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type CreateGameModeUseCase struct {
	GameModeWriter out.GameModeWriter
	GameModeReader out.GameModeReader
}

func NewCreateGameModeUseCase(gameModeWriter out.GameModeWriter, gameModeReader out.GameModeReader) in.CreateGameModeCommand {
	return &CreateGameModeUseCase{
		GameModeWriter: gameModeWriter,
		GameModeReader: gameModeReader,
	}
}

func InjectCreateGameMode(c container.Container) error {
	c.Singleton(func(gameModeWriter out.GameModeWriter, gameModeReader out.GameModeReader) (in.CreateGameModeCommand, error) {
		return NewCreateGameModeUseCase(gameModeWriter, gameModeReader), nil
	})
	return nil
}

func (usecase *CreateGameModeUseCase) Execute(ctx context.Context, gameMode *game_entities.GameMode) (*game_entities.GameMode, error) {
	// Validate game mode data
	if err := validateGameMode(gameMode); err != nil {
		slog.ErrorContext(ctx, "game mode validation failed", "error", err)
		return nil, fmt.Errorf("invalid game mode data: %w", err)
	}

	// Check if a game mode with the same name already exists for this game
	existingGameModes, err := usecase.GameModeReader.Search(ctx, nil)
	if err == nil {
		for _, existing := range existingGameModes {
			if existing.GameID == gameMode.GameID && existing.Name == gameMode.Name {
				slog.WarnContext(ctx, "game mode with same name already exists for game", "name", gameMode.Name, "game_id", gameMode.GameID)
				return nil, fmt.Errorf("game mode with name '%s' already exists for this game", gameMode.Name)
			}
		}
	} else {
		// If there's an error in the search, just log but continue
		slog.WarnContext(ctx, "failed to search existing game modes", "error", err)
	}

	// Create base entity
	resourceOwner := common.GetResourceOwner(ctx)
	baseEntity := common.NewEntity(resourceOwner)
	gameMode.BaseEntity = baseEntity

	// Audit log
	slog.InfoContext(ctx, "creating new game mode", "name", gameMode.Name, "game_id", gameMode.GameID, "user_id", resourceOwner.UserID)

	// Create the game mode (repository will create the ID automatically)
	createdGameMode, err := usecase.GameModeWriter.Create(ctx, gameMode)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create game mode", "error", err, "name", gameMode.Name)
		return nil, fmt.Errorf("failed to create game mode: %w", err)
	}

	slog.InfoContext(ctx, "game mode created successfully", "game_mode_id", createdGameMode.ID, "name", createdGameMode.Name)

	return createdGameMode, nil
}

// validateGameMode validates game mode data before creating or updating
func validateGameMode(gameMode *game_entities.GameMode) error {
	if gameMode.Name == "" {
		return errors.New("game mode name is required")
	}

	if len(gameMode.Name) > 100 {
		return errors.New("game mode name must be 100 characters or less")
	}

	if gameMode.GameID == uuid.Nil {
		return errors.New("game_id is required and must be a valid UUID")
	}

	return nil
}
