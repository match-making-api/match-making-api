package entities

import (
	"time"

	"github.com/google/uuid"
)

// LobbyStatus represents the current state of a lobby
type LobbyStatus string

const (
	LobbyStatusOpen       LobbyStatus = "open"        // Accepting players
	LobbyStatusReadyCheck LobbyStatus = "ready_check" // Countdown active
	LobbyStatusStarting   LobbyStatus = "starting"    // Creating match
	LobbyStatusStarted    LobbyStatus = "started"     // Match in progress
	LobbyStatusCancelled  LobbyStatus = "cancelled"   // Lobby cancelled
	LobbyStatusCompleted  LobbyStatus = "completed"   // Match finished
)

// LobbyVisibility controls who can see and join the lobby
type LobbyVisibility string

const (
	LobbyVisibilityPublic     LobbyVisibility = "public"      // Anyone can see and join
	LobbyVisibilityPrivate    LobbyVisibility = "private"     // Invite only, hidden from browse
	LobbyVisibilityMatchmaking LobbyVisibility = "matchmaking" // System managed, limited info shown
	LobbyVisibilityFriends    LobbyVisibility = "friends"     // Only friends can see/join
)

// LobbyType defines the type of lobby
type LobbyType string

const (
	LobbyTypeCustom      LobbyType = "custom"      // Player-created custom lobby
	LobbyTypeRanked      LobbyType = "ranked"      // Ranked competitive
	LobbyTypeCasual      LobbyType = "casual"      // Casual unranked
	LobbyTypeTournament  LobbyType = "tournament"  // Tournament match
	LobbyTypePractice    LobbyType = "practice"    // Practice/scrimmage
)

// PlayerSlot represents a player position in the lobby
type PlayerSlot struct {
	SlotNumber  int       `json:"slot_number" bson:"slot_number"`
	PlayerID    *uuid.UUID `json:"player_id,omitempty" bson:"player_id,omitempty"`
	PlayerName  string    `json:"player_name,omitempty" bson:"player_name,omitempty"`
	IsReady     bool      `json:"is_ready" bson:"is_ready"`
	JoinedAt    time.Time `json:"joined_at" bson:"joined_at"`
	MMR         int       `json:"mmr,omitempty" bson:"mmr,omitempty"`
	Rank        string    `json:"rank,omitempty" bson:"rank,omitempty"`
	Team        int       `json:"team" bson:"team"` // 0=unassigned, 1=team1, 2=team2
	IsSpectator bool      `json:"is_spectator" bson:"is_spectator"`
}

// SkillRange defines MMR boundaries for matchmaking
type SkillRange struct {
	MinMMR int `json:"min_mmr" bson:"min_mmr"`
	MaxMMR int `json:"max_mmr" bson:"max_mmr"`
}

// PrizePoolConfig holds entry fee and distribution settings
type PrizePoolConfig struct {
	EntryFeeCents    int    `json:"entry_fee_cents" bson:"entry_fee_cents"`
	PrizePoolID      string `json:"prize_pool_id,omitempty" bson:"prize_pool_id,omitempty"`
	DistributionRule string `json:"distribution_rule" bson:"distribution_rule"` // winner_takes_all, top_3, etc.
}

// QueueStats holds information about players waiting (for matchmaking visibility)
type QueueStats struct {
	PlayersWaiting    int           `json:"players_waiting" bson:"players_waiting"`
	AverageWaitTime   time.Duration `json:"average_wait_time" bson:"average_wait_time"`
	EstimatedWaitTime time.Duration `json:"estimated_wait_time" bson:"estimated_wait_time"`
}

// Lobby represents a game lobby for matchmaking
type Lobby struct {
	// Identity
	ID        uuid.UUID `json:"id" bson:"_id"`
	TenantID  uuid.UUID `json:"tenant_id" bson:"tenant_id"`
	ClientID  uuid.UUID `json:"client_id" bson:"client_id"`
	CreatorID uuid.UUID `json:"creator_id" bson:"creator_id"`

	// Game Configuration
	GameID   string   `json:"game_id" bson:"game_id"`     // cs2, valorant, etc.
	GameMode string   `json:"game_mode" bson:"game_mode"` // competitive, casual, etc.
	MapPool  []string `json:"map_pool,omitempty" bson:"map_pool,omitempty"`
	Region   string   `json:"region" bson:"region"` // na, eu, br, sea, etc.

	// Lobby Settings
	Name        string          `json:"name" bson:"name"`
	Description string          `json:"description,omitempty" bson:"description,omitempty"`
	Type        LobbyType       `json:"type" bson:"type"`
	Visibility  LobbyVisibility `json:"visibility" bson:"visibility"`
	IsFeatured  bool            `json:"is_featured" bson:"is_featured"` // For homepage display
	Tags        []string        `json:"tags,omitempty" bson:"tags,omitempty"`

	// Player Configuration
	MaxPlayers        int  `json:"max_players" bson:"max_players"`
	MinPlayers        int  `json:"min_players" bson:"min_players"`
	RequiresReadyCheck bool `json:"requires_ready_check" bson:"requires_ready_check"`
	AllowSpectators   bool `json:"allow_spectators" bson:"allow_spectators"`
	AllowCrossPlatform bool `json:"allow_cross_platform" bson:"allow_cross_platform"`

	// Players
	PlayerSlots  []PlayerSlot `json:"player_slots" bson:"player_slots"`
	SpectatorIDs []uuid.UUID  `json:"spectator_ids,omitempty" bson:"spectator_ids,omitempty"`

	// Skill/Ranking
	SkillRange *SkillRange `json:"skill_range,omitempty" bson:"skill_range,omitempty"`
	MaxPing    int         `json:"max_ping,omitempty" bson:"max_ping,omitempty"`

	// Prize Pool
	PrizePool *PrizePoolConfig `json:"prize_pool,omitempty" bson:"prize_pool,omitempty"`

	// Status & Timing
	Status      LobbyStatus `json:"status" bson:"status"`
	CreatedAt   time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" bson:"updated_at"`
	ExpiresAt   time.Time   `json:"expires_at" bson:"expires_at"`
	StartedAt   *time.Time  `json:"started_at,omitempty" bson:"started_at,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty" bson:"completed_at,omitempty"`

	// Match Result
	MatchID        *uuid.UUID  `json:"match_id,omitempty" bson:"match_id,omitempty"`
	WinnerPlayerIDs []uuid.UUID `json:"winner_player_ids,omitempty" bson:"winner_player_ids,omitempty"`

	// Queue Stats (for matchmaking type lobbies)
	QueueStats *QueueStats `json:"queue_stats,omitempty" bson:"queue_stats,omitempty"`

	// Metadata for extensibility
	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// GetCurrentPlayerCount returns the number of players in the lobby
func (l *Lobby) GetCurrentPlayerCount() int {
	count := 0
	for _, slot := range l.PlayerSlots {
		if slot.PlayerID != nil && !slot.IsSpectator {
			count++
		}
	}
	return count
}

// GetReadyPlayerCount returns the number of ready players
func (l *Lobby) GetReadyPlayerCount() int {
	count := 0
	for _, slot := range l.PlayerSlots {
		if slot.PlayerID != nil && slot.IsReady && !slot.IsSpectator {
			count++
		}
	}
	return count
}

// IsFull returns true if the lobby has max players
func (l *Lobby) IsFull() bool {
	return l.GetCurrentPlayerCount() >= l.MaxPlayers
}

// CanStart returns true if lobby has minimum players ready
func (l *Lobby) CanStart() bool {
	readyCount := l.GetReadyPlayerCount()
	return readyCount >= l.MinPlayers
}

// HasPlayer checks if a player is in the lobby
func (l *Lobby) HasPlayer(playerID uuid.UUID) bool {
	for _, slot := range l.PlayerSlots {
		if slot.PlayerID != nil && *slot.PlayerID == playerID {
			return true
		}
	}
	return false
}

// GetPublicView returns a sanitized view for public/matchmaking lobbies
func (l *Lobby) GetPublicView() *Lobby {
	public := *l
	
	// For matchmaking lobbies, hide player details
	if l.Visibility == LobbyVisibilityMatchmaking {
		for i := range public.PlayerSlots {
			public.PlayerSlots[i].PlayerName = ""
			public.PlayerSlots[i].MMR = 0
		}
	}
	
	return &public
}
