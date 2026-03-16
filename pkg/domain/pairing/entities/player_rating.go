package entities

import (
	"time"

	"github.com/google/uuid"
)

// PlayerRating represents a player's MMR/rating for a game, with resource ownership.
// Stored per player_id + game_id + tenant_id + client_id; used for matchmaking and leaderboards.
type PlayerRating struct {
	PlayerID        string    `json:"player_id" bson:"player_id"`
	GameID          string    `json:"game_id" bson:"game_id"` // Empty for global rating
	TenantID        string    `json:"tenant_id" bson:"tenant_id"`
	ClientID        string    `json:"client_id" bson:"client_id"`
	ResourceOwnerID string    `json:"resource_owner_id" bson:"resource_owner_id"`
	MMR             int32     `json:"mmr" bson:"mmr"`
	UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`
}

// RatingAuditEntry records a rating change for audit trail.
type RatingAuditEntry struct {
	MatchID          uuid.UUID `json:"match_id" bson:"match_id"`
	PlayerID         string    `json:"player_id" bson:"player_id"`
	MMRBefore        int32     `json:"mmr_before" bson:"mmr_before"`
	MMRAfter         int32     `json:"mmr_after" bson:"mmr_after"`
	Delta            int32     `json:"delta" bson:"delta"`
	AlgorithmVersion string    `json:"algorithm_version" bson:"algorithm_version"`
	Reason           string    `json:"reason" bson:"reason"`
	UpdatedBy        string    `json:"updated_by" bson:"updated_by"`
	UpdatedAt        time.Time `json:"updated_at" bson:"updated_at"`
}
