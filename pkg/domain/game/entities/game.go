package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

type Game struct {
	common.BaseEntity
	ID          uuid.UUID `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`

	// Configurações de equipe
	MinPlayersPerTeam int `json:"min_players_per_team" bson:"min_players_per_team"`
	MaxPlayersPerTeam int `json:"max_players_per_team" bson:"max_players_per_team"`
	NumberOfTeams     int `json:"number_of_teams" bson:"number_of_teams"`

	// Configurações de partida
	MaxDuration     time.Duration `json:"max_duration" bson:"max_duration"`
	AllowSpectators bool          `json:"allow_spectators" bson:"allow_spectators"`

	// Configurações de matchmaking
	SkillBasedMatching bool     `json:"skill_based_matching" bson:"skill_based_matching"`
	AllowedRegions     []string `json:"allowed_regions" bson:"allowed_regions"`

	// Configurações de jogo
	GameModes   []string          `json:"game_modes" bson:"game_modes"`
	MapPool     []string          `json:"map_pool" bson:"map_pool"`
	CustomRules map[string]string `json:"custom_rules" bson:"custom_rules"`

	// Metadados
	Enabled bool `json:"enabled" bson:"enabled"`
}
