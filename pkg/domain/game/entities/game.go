package entities

import (
	"time"

	"github.com/leet-gaming/match-making-api/pkg/common"
)

// Game struct represents a game with its properties and settings
type Game struct {
	common.BaseEntity
	Name        string `json:"name" bson:"name"`               // Name of the game
	Description string `json:"description" bson:"description"` // Description of the game

	// Team settings
	MinPlayersPerTeam int `json:"min_players_per_team" bson:"min_players_per_team"` // Minimum number of players per team
	MaxPlayersPerTeam int `json:"max_players_per_team" bson:"max_players_per_team"` // Maximum number of players per team
	NumberOfTeams     int `json:"number_of_teams" bson:"number_of_teams"`           // Number of teams in the game

	// Match settings
	MaxDuration     time.Duration `json:"max_duration" bson:"max_duration"`         // Maximum duration of a match
	AllowSpectators bool          `json:"allow_spectators" bson:"allow_spectators"` // Allow spectators to join the game

	// Matchmaking settings
	SkillBasedMatching bool     `json:"skill_based_matching" bson:"skill_based_matching"` // Enable skill-based matchmaking
	AllowedRegions     []string `json:"allowed_regions" bson:"allowed_regions"`           // Allowed regions for matchmaking

	// Game settings
	GameModes   []string          `json:"game_modes" bson:"game_modes"`     // Game modes available in the game
	MapPool     []string          `json:"map_pool" bson:"map_pool"`         // Map pool for the game
	CustomRules map[string]string `json:"custom_rules" bson:"custom_rules"` // Custom rules for the game

	// Metadata
	Enabled bool `json:"enabled" bson:"enabled"` // Enable/disable the game
}
