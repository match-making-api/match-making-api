package pairing_out

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
)

// ServerAllocationQueueStore defines the contract for the queue of matches
// waiting for game server allocation. FIFO ordering per game_id:region.
// Used when no server is available; matches are enqueued until a server is free.
type ServerAllocationQueueStore interface {
	// Enqueue adds a match to the queue for the given game/region.
	// Ordering is FIFO per game_id:region.
	Enqueue(ctx context.Context, entry *entities.ServerAllocationQueueEntry) error

	// DequeueNext removes and returns the next match for the given game/region.
	// Returns nil if queue is empty.
	DequeueNext(ctx context.Context, gameID uuid.UUID, region string) (*entities.ServerAllocationQueueEntry, error)

	// PeekNext returns the next match without removing it.
	PeekNext(ctx context.Context, gameID uuid.UUID, region string) (*entities.ServerAllocationQueueEntry, error)

	// Remove removes a specific match from the queue (e.g. on timeout/abandon).
	Remove(ctx context.Context, matchID uuid.UUID) error

	// GetByMatchID returns the entry for a match if it exists in any queue.
	GetByMatchID(ctx context.Context, matchID uuid.UUID) (*entities.ServerAllocationQueueEntry, error)

	// ListStale returns entries enqueued before the given cutoff time (for timeout processing).
	ListStale(ctx context.Context, cutoff time.Time) ([]*entities.ServerAllocationQueueEntry, error)
}
