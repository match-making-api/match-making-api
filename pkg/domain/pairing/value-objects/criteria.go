package value_objects

import (
	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
)

type Criteria struct {
	TenantID *uuid.UUID
	ClientID *uuid.UUID
	// Game     *lobbies_entities.Game // TODO: ideate modes/rank/rating_range (game -> tenant + client)
	Schedule *schedule_entities.Schedule
	Region   *game_entities.Region

	PairSize int //Edges    map[int]int
	// MinParties int
	// MaxParties int

	// Matchmaking-specific criteria - Entity references for domain consistency
	GameID            *uuid.UUID  `json:"game_id,omitempty" bson:"game_id,omitempty"`
	GameModeID        *uuid.UUID  `json:"game_mode_id,omitempty" bson:"game_mode_id,omitempty"`
	MapPreferences    []string    `json:"map_preferences,omitempty" bson:"map_preferences,omitempty"`
	SkillRange        *SkillRange `json:"skill_range,omitempty" bson:"skill_range,omitempty"`
	MaxPing           int         `json:"max_ping,omitempty" bson:"max_ping,omitempty"`
	AllowCrossPlatform bool       `json:"allow_cross_platform,omitempty" bson:"allow_cross_platform,omitempty"`
	Tier              string      `json:"tier,omitempty" bson:"tier,omitempty"`
	PriorityBoost     bool        `json:"priority_boost,omitempty" bson:"priority_boost,omitempty"`
}

// SkillRange defines acceptable skill level range for matchmaking
type SkillRange struct {
	MinMMR int `json:"min_mmr" bson:"min_mmr"`
	MaxMMR int `json:"max_mmr" bson:"max_mmr"`
}
