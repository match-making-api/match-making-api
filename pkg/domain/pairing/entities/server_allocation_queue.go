package entities

import (
	"time"

	"github.com/google/uuid"
)

// ServerAllocationQueueEntry represents a match waiting for game server allocation.
// Stored in FIFO order per game_id:region for fair allocation when servers become available.
type ServerAllocationQueueEntry struct {
	MatchID         uuid.UUID   `json:"match_id"`
	GameID          uuid.UUID   `json:"game_id"`
	Region          string      `json:"region"`
	PlayerIDs       []uuid.UUID `json:"player_ids"`
	TenantID        string      `json:"tenant_id"`
	ClientID        string      `json:"client_id"`
	ResourceOwnerID string      `json:"resource_owner_id"`
	EnqueuedAt      time.Time   `json:"enqueued_at"`
}
