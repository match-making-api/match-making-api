package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// PlayerRating represents a player's MMR/rating for a game, with resource ownership.
// Stored per player_id + game_id + tenant_id + client_id; used for matchmaking and leaderboards.
type PlayerRating struct {
	PlayerID string `json:"player_id" bson:"player_id"`
	GameID   string `json:"game_id" bson:"game_id"` // Empty for global rating
	// Flat fields kept for dual-write / query keys during migration (2508-001).
	TenantID        string `json:"tenant_id" bson:"tenant_id"`
	ClientID        string `json:"client_id" bson:"client_id"`
	ResourceOwnerID string `json:"resource_owner_id" bson:"resource_owner_id"`
	// ResourceOwner is the canonical nested ownership document.
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	MMR           int32                `json:"mmr" bson:"mmr"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`
}

// EnsureResourceOwner syncs nested ResourceOwner ↔ flat string IDs.
func (r *PlayerRating) EnsureResourceOwner() {
	if r.ResourceOwner.TenantID == uuid.Nil && r.TenantID != "" {
		r.ResourceOwner = common.ResourceOwnerFromFlat(r.TenantID, r.ClientID, "", r.ResourceOwnerID)
	}
	t, c, _, u := r.ResourceOwner.FlatStrings()
	if r.TenantID == "" {
		r.TenantID = t
	}
	if r.ClientID == "" {
		r.ClientID = c
	}
	if r.ResourceOwnerID == "" {
		r.ResourceOwnerID = u
	}
}

// ValidateOwnership requires tenant + client after sync.
func (r *PlayerRating) ValidateOwnership() error {
	r.EnsureResourceOwner()
	return r.ResourceOwner.ValidateTenantClient()
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
