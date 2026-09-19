package entities

// LobbyParams are mode-specific settings used when creating a match/lobby
// (match-making-api → replay-api exchange; Refs 2508-005).
type LobbyParams struct {
	// PartySize is the expected number of players per party (solo = 1).
	PartySize int `json:"party_size" bson:"party_size"`
	// MaxPlayers is the lobby capacity (all teams).
	MaxPlayers int `json:"max_players" bson:"max_players"`
	// TeamCount is the number of teams in the match.
	TeamCount int `json:"team_count" bson:"team_count"`
	// MapPool overrides or narrows the game-level map pool for this mode.
	MapPool []string `json:"map_pool,omitempty" bson:"map_pool,omitempty"`
	// AllowedRegions restricts matchmaking regions for this mode (empty = use game).
	AllowedRegions []string `json:"allowed_regions,omitempty" bson:"allowed_regions,omitempty"`
	// CustomRules mode-level rules merged over game CustomRules.
	CustomRules map[string]string `json:"custom_rules,omitempty" bson:"custom_rules,omitempty"`
	// ReadyCheckSeconds is the readiness window; 0 = use platform default.
	ReadyCheckSeconds int `json:"ready_check_seconds,omitempty" bson:"ready_check_seconds,omitempty"`
}

// GameConfiguration is the resolved view of game rules + lobby params for a
// tenant/client-scoped game mode (source of truth in match-making Mongo).
type GameConfiguration struct {
	GameID     string            `json:"game_id"`
	GameModeID string            `json:"game_mode_id"`
	GameName   string            `json:"game_name"`
	ModeName   string            `json:"mode_name"`
	TenantID   string            `json:"tenant_id"`
	ClientID   string            `json:"client_id"`
	Enabled    bool              `json:"enabled"`
	MapPool    []string          `json:"map_pool"`
	Regions    []string          `json:"regions"`
	Rules      map[string]string `json:"rules"`
	Lobby      LobbyParams       `json:"lobby"`
}
