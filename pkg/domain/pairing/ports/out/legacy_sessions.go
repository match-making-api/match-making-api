package pairing_out

import (
	"context"

	"github.com/google/uuid"
)

// LegacySessionRef is a minimal projection of a replay-api matchmaking session
// used for dual-write consistency checks (2509-002).
type LegacySessionRef struct {
	PlayerID uuid.UUID
	Status   string
	MatchID  *uuid.UUID
}

// LegacySessionReader reads legacy matchmaking sessions (Mongo matchmaking_sessions).
// Owned by replay-api; match-making-api reads for monitoring only.
type LegacySessionReader interface {
	// ListByStatuses returns non-expired sessions with the given statuses.
	ListByStatuses(ctx context.Context, statuses []string) ([]LegacySessionRef, error)
}
