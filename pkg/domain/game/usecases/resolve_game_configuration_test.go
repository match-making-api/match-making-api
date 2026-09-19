package usecases

import (
	"testing"

	gofrsuuid "github.com/gofrs/uuid"
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
)

func TestMergeGameConfiguration_ModeOverrides(t *testing.T) {
	gameID := gofrsuuid.Must(gofrsuuid.NewV4())

	game := &game_entities.Game{
		BaseEntity:        common.BaseEntity{ID: uuid.MustParse(gameID.String())},
		Name:              "CS",
		MaxPlayersPerTeam: 5,
		NumberOfTeams:     2,
		MapPool:           []string{"dust2", "mirage"},
		AllowedRegions:    []string{"br", "us"},
		CustomRules:       map[string]string{"ot": "mr3"},
		Enabled:           true,
	}
	mode := &game_entities.GameMode{
		BaseEntity: common.BaseEntity{
			ID: uuid.New(),
			ResourceOwner: common.ResourceOwner{
				TenantID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				ClientID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			},
		},
		GameID: gameID,
		Name:   "5v5",
		Lobby: game_entities.LobbyParams{
			PartySize:  2,
			MaxPlayers: 10,
			TeamCount:  2,
			MapPool:    []string{"inferno"},
			CustomRules: map[string]string{
				"knife": "enabled",
			},
		},
	}

	cfg := MergeGameConfiguration(game, mode)
	if cfg.Lobby.PartySize != 2 || cfg.Lobby.MaxPlayers != 10 {
		t.Fatalf("lobby: %+v", cfg.Lobby)
	}
	if len(cfg.MapPool) != 1 || cfg.MapPool[0] != "inferno" {
		t.Fatalf("maps: %v", cfg.MapPool)
	}
	if cfg.Rules["ot"] != "mr3" || cfg.Rules["knife"] != "enabled" {
		t.Fatalf("rules: %v", cfg.Rules)
	}
	if cfg.TenantID == "" || cfg.ClientID == "" {
		t.Fatal("tenant/client scoping missing")
	}
}

func TestMergeGameConfiguration_DefaultsFromGame(t *testing.T) {
	gid := gofrsuuid.Must(gofrsuuid.NewV4())
	game := &game_entities.Game{
		Name:              "Game",
		MaxPlayersPerTeam: 1,
		NumberOfTeams:     2,
		MapPool:           []string{"a"},
		AllowedRegions:    []string{"eu"},
		Enabled:           true,
	}
	mode := &game_entities.GameMode{
		BaseEntity: common.BaseEntity{ID: uuid.New()},
		GameID:     gid,
		Name:       "1v1",
	}
	cfg := MergeGameConfiguration(game, mode)
	if cfg.Lobby.MaxPlayers != 2 || cfg.Lobby.PartySize != 1 {
		t.Fatalf("expected defaults, got %+v", cfg.Lobby)
	}
	if len(cfg.Regions) != 1 || cfg.Regions[0] != "eu" {
		t.Fatalf("regions: %v", cfg.Regions)
	}
}
