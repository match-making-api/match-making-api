package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type CreateGameUseCase struct {
	GameWriter out.GameWriter
	GameReader out.GameReader
}

func NewCreateGameUseCase(gameWriter out.GameWriter, gameReader out.GameReader) in.CreateGameCommand {
	return &CreateGameUseCase{
		GameWriter: gameWriter,
		GameReader: gameReader,
	}
}

func InjectCreateGame(c container.Container) error {
	c.Singleton(func(gameWriter out.GameWriter, gameReader out.GameReader) (in.CreateGameCommand, error) {
		return NewCreateGameUseCase(gameWriter, gameReader), nil
	})
	return nil
}

func (usecase *CreateGameUseCase) Execute(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error) {
	// Validate game data
	if err := validateGame(game); err != nil {
		slog.ErrorContext(ctx, "game validation failed", "error", err)
		return nil, fmt.Errorf("invalid game data: %w", err)
	}

	// Check if a game with the same name already exists
	existingGames, err := usecase.GameReader.Search(ctx, nil)
	if err == nil {
		for _, existing := range existingGames {
			if existing.Name == game.Name {
				slog.WarnContext(ctx, "game with same name already exists", "name", game.Name)
				return nil, fmt.Errorf("game with name '%s' already exists", game.Name)
			}
		}
	} else {
		// If there's an error in the search, just log but continue (may be no games exist yet)
		slog.WarnContext(ctx, "failed to search existing games", "error", err)
	}

	// Create base entity
	resourceOwner := common.GetResourceOwner(ctx)
	baseEntity := common.NewEntity(resourceOwner)
	game.BaseEntity = baseEntity

	// Set default values if necessary
	if game.MaxDuration == 0 {
		game.MaxDuration = 30 * time.Minute
	}
	// By default, new games are enabled
	game.Enabled = true

	// Audit log
	slog.InfoContext(ctx, "creating new game", "name", game.Name, "user_id", resourceOwner.UserID)

	// Create the game (repository will create the ID automatically)
	createdGame, err := usecase.GameWriter.Create(ctx, game)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create game", "error", err, "name", game.Name)
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	slog.InfoContext(ctx, "game created successfully", "game_id", createdGame.ID, "name", createdGame.Name)

	return createdGame, nil
}

// validateGame validates game data before creating or updating
func validateGame(game *game_entities.Game) error {
	if game.Name == "" {
		return errors.New("game name is required")
	}

	if len(game.Name) > 100 {
		return errors.New("game name must be 100 characters or less")
	}

	if game.MinPlayersPerTeam <= 0 {
		return errors.New("min_players_per_team must be greater than 0")
	}

	if game.MaxPlayersPerTeam <= 0 {
		return errors.New("max_players_per_team must be greater than 0")
	}

	if game.MinPlayersPerTeam > game.MaxPlayersPerTeam {
		return errors.New("min_players_per_team cannot be greater than max_players_per_team")
	}

	if game.NumberOfTeams <= 0 {
		return errors.New("number_of_teams must be greater than 0")
	}

	if game.MaxDuration < 0 {
		return errors.New("max_duration cannot be negative")
	}

	return nil
}
