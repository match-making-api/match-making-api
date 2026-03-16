package common

// LobbyStats represents aggregate statistics about lobbies
type LobbyStats struct {
	TotalActiveLobbies   int            `json:"total_active_lobbies"`
	TotalPlayersWaiting  int            `json:"total_players_waiting"`
	ByGame               map[string]int `json:"by_game"`
	ByRegion             map[string]int `json:"by_region"`
	ByMode               map[string]int `json:"by_mode"`
	AvgFillTimeSeconds   float64        `json:"avg_fill_time_seconds"`
	AvgPlayersPerLobby   float64        `json:"avg_players_per_lobby"`
}
