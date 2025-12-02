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

func InjectUpdateGame(c container.Container) error {
	c.Singleton(func(gameWriter out.GameWriter, gameReader out.GameReader) (in.UpdateGameCommand, error) {
		return NewUpdateGameUseCase(gameWriter, gameReader), nil
	})
	return nil
}

func (usecase *UpdateGameUseCase) Execute(ctx context.Context, id uuid.UUID, game *game_entities.Game) (*game_entities.Game, error) {
	// Get existing game
	existingGame, err := usecase.GameReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "game not found", "game_id", id, "error", err)
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// Validate game data
	if err := validateGame(game); err != nil {
		slog.ErrorContext(ctx, "game validation failed", "error", err, "game_id", id)
		return nil, fmt.Errorf("invalid game data: %w", err)
	}

	// Check if another game with the same name already exists (except the current one)
	existingGames, err := usecase.GameReader.Search(ctx, nil)
	if err == nil {
		for _, existing := range existingGames {
			if existing.ID != id && existing.Name == game.Name {
				slog.WarnContext(ctx, "game with same name already exists", "name", game.Name, "existing_id", existing.ID)
				return nil, fmt.Errorf("game with name '%s' already exists", game.Name)
			}
		}
	}

	// Audit log before update
	resourceOwner := common.GetResourceOwner(ctx)
	slog.InfoContext(ctx, "updating game", "game_id", id, "name", game.Name, "user_id", resourceOwner.UserID)

	// Update all fields
	existingGame.Name = game.Name
	existingGame.Description = game.Description
	existingGame.MinPlayersPerTeam = game.MinPlayersPerTeam
	existingGame.MaxPlayersPerTeam = game.MaxPlayersPerTeam
	existingGame.NumberOfTeams = game.NumberOfTeams
	existingGame.MaxDuration = game.MaxDuration
	existingGame.AllowSpectators = game.AllowSpectators
	existingGame.SkillBasedMatching = game.SkillBasedMatching
	existingGame.AllowedRegions = game.AllowedRegions
	existingGame.GameModes = game.GameModes
	existingGame.MapPool = game.MapPool
	existingGame.CustomRules = game.CustomRules
	existingGame.Enabled = game.Enabled
	existingGame.UpdatedAt = time.Now()

	// Update the game
	updatedGame, err := usecase.GameWriter.Update(ctx, existingGame)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update game", "error", err, "game_id", id)
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	slog.InfoContext(ctx, "game updated successfully", "game_id", updatedGame.ID, "name", updatedGame.Name)

	return updatedGame, nil
}
