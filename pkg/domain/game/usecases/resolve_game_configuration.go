package usecases

import (
	"context"
	"fmt"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

// ResolveGameConfigurationQuery resolves game rules + lobby params for a mode.
type ResolveGameConfigurationQuery interface {
	Execute(ctx context.Context, gameModeID uuid.UUID) (*game_entities.GameConfiguration, error)
}

// ResolveGameConfigurationUseCase loads GameMode + Game and merges lobby params.
type ResolveGameConfigurationUseCase struct {
	GameModeReader out.GameModeReader
	GameReader     out.GameReader
}

// NewResolveGameConfigurationUseCase constructs the use case.
func NewResolveGameConfigurationUseCase(modeReader out.GameModeReader, gameReader out.GameReader) *ResolveGameConfigurationUseCase {
	return &ResolveGameConfigurationUseCase{GameModeReader: modeReader, GameReader: gameReader}
}

// InjectResolveGameConfiguration registers the query in the IoC container.
func InjectResolveGameConfiguration(c container.Container) error {
	return c.Singleton(func(modeReader out.GameModeReader, gameReader out.GameReader) (ResolveGameConfigurationQuery, error) {
		return NewResolveGameConfigurationUseCase(modeReader, gameReader), nil
	})
}

// Execute returns the tenant/client-scoped configuration for lobby/match creation.
func (uc *ResolveGameConfigurationUseCase) Execute(ctx context.Context, gameModeID uuid.UUID) (*game_entities.GameConfiguration, error) {
	mode, err := uc.GameModeReader.GetByID(ctx, gameModeID)
	if err != nil {
		return nil, fmt.Errorf("game mode: %w", err)
	}
	if mode == nil {
		return nil, fmt.Errorf("game mode not found")
	}

	gameID, err := uuid.Parse(mode.GameID.String())
	if err != nil {
		return nil, fmt.Errorf("invalid game_id on mode: %w", err)
	}

	game, err := uc.GameReader.GetByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("game: %w", err)
	}
	if game == nil {
		return nil, fmt.Errorf("game not found for mode")
	}

	return MergeGameConfiguration(game, mode), nil
}

// MergeGameConfiguration builds LobbyCreation-ready config from Game + GameMode.
// Mode lobby fields win over game defaults when set (party size / max players > 0).
func MergeGameConfiguration(game *game_entities.Game, mode *game_entities.GameMode) *game_entities.GameConfiguration {
	lobby := mode.Lobby
	if lobby.PartySize <= 0 {
		lobby.PartySize = 1
	}
	if lobby.MaxPlayers <= 0 {
		perTeam := game.MaxPlayersPerTeam
		if perTeam <= 0 {
			perTeam = 1
		}
		teams := game.NumberOfTeams
		if teams <= 0 {
			teams = 2
		}
		lobby.MaxPlayers = perTeam * teams
		lobby.TeamCount = teams
	}
	if lobby.TeamCount <= 0 {
		lobby.TeamCount = game.NumberOfTeams
		if lobby.TeamCount <= 0 {
			lobby.TeamCount = 2
		}
	}

	maps := lobby.MapPool
	if len(maps) == 0 {
		maps = append([]string(nil), game.MapPool...)
	}
	regions := lobby.AllowedRegions
	if len(regions) == 0 {
		regions = append([]string(nil), game.AllowedRegions...)
	}
	rules := map[string]string{}
	for k, v := range game.CustomRules {
		rules[k] = v
	}
	for k, v := range lobby.CustomRules {
		rules[k] = v
	}

	return &game_entities.GameConfiguration{
		GameID:     mode.GameID.String(),
		GameModeID: mode.ID.String(),
		GameName:   game.Name,
		ModeName:   mode.Name,
		TenantID:   mode.ResourceOwner.TenantID.String(),
		ClientID:   mode.ResourceOwner.ClientID.String(),
		Enabled:    game.Enabled,
		MapPool:    maps,
		Regions:    regions,
		Rules:      rules,
		Lobby:      lobby,
	}
}
