package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// MatchResult represents persisted match results with resource ownership.
// Stored after consuming MatchCompleted; used for audit and access control.
type MatchResult struct {
	MatchID         uuid.UUID `json:"match_id" bson:"match_id"`
	PlayerIDs       []string  `json:"player_ids" bson:"player_ids"`
	WinnerTeamID    string    `json:"winner_team_id" bson:"winner_team_id"`
	IsDraw          bool      `json:"is_draw" bson:"is_draw"`
	CompletedAtMs   int64     `json:"completed_at_epoch_ms" bson:"completed_at_epoch_ms"`
	CalculatedAtMs  int64     `json:"calculated_at_epoch_ms" bson:"calculated_at_epoch_ms"`
	TenantID        string    `json:"tenant_id" bson:"tenant_id"`
	ClientID        string    `json:"client_id" bson:"client_id"`
	ResourceOwnerID string    `json:"resource_owner_id" bson:"resource_owner_id"`
	// ResourceOwner is the canonical nested ownership document (2508-001).
	ResourceOwner  common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	SourceEventID  string               `json:"source_event_id" bson:"source_event_id"`
	CalculatedAt   time.Time            `json:"calculated_at" bson:"calculated_at"`
}

// EnsureResourceOwner syncs nested ResourceOwner ↔ flat string IDs.
func (m *MatchResult) EnsureResourceOwner() {
	if m.ResourceOwner.TenantID == uuid.Nil && m.TenantID != "" {
		m.ResourceOwner = common.ResourceOwnerFromFlat(m.TenantID, m.ClientID, "", m.ResourceOwnerID)
	}
	t, c, _, u := m.ResourceOwner.FlatStrings()
	if m.TenantID == "" {
		m.TenantID = t
	}
	if m.ClientID == "" {
		m.ClientID = c
	}
	if m.ResourceOwnerID == "" {
		m.ResourceOwnerID = u
	}
}

// ValidateOwnership requires tenant + client after sync.
func (m *MatchResult) ValidateOwnership() error {
	m.EnsureResourceOwner()
	return m.ResourceOwner.ValidateTenantClient()
}
