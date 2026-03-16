package entities

import (
	"time"

	"github.com/google/uuid"
)

// MatchResult represents persisted match results with resource ownership.
// Stored after consuming MatchCompleted; used for audit and access control.
type MatchResult struct {
	MatchID          uuid.UUID `json:"match_id" bson:"match_id"`
	PlayerIDs        []string  `json:"player_ids" bson:"player_ids"`
	WinnerTeamID     string    `json:"winner_team_id" bson:"winner_team_id"`
	IsDraw           bool      `json:"is_draw" bson:"is_draw"`
	CompletedAtMs    int64     `json:"completed_at_epoch_ms" bson:"completed_at_epoch_ms"`
	CalculatedAtMs   int64     `json:"calculated_at_epoch_ms" bson:"calculated_at_epoch_ms"`
	TenantID         string    `json:"tenant_id" bson:"tenant_id"`
	ClientID         string    `json:"client_id" bson:"client_id"`
	ResourceOwnerID  string    `json:"resource_owner_id" bson:"resource_owner_id"`
	SourceEventID    string    `json:"source_event_id" bson:"source_event_id"` // MatchCompleted event_id for audit
	CalculatedAt     time.Time `json:"calculated_at" bson:"calculated_at"`
}
