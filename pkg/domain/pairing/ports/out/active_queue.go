package pairing_out

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
)

// ActiveQueueStore defines the contract for tracking players actively waiting
// in matchmaking queues. Implementations may be in-memory (for testing) or
// backed by a distributed store like Redis/Dragonfly (for production).
type ActiveQueueStore interface {
	// Register adds or updates a player in the active queue.
	// If the player is already registered, their position is updated
	// but the original JoinedAt time is preserved.
	Register(ctx context.Context, entry *entities.ActiveQueueEntry) error

	// Remove removes a player from the active queue (e.g. matched or left).
	Remove(ctx context.Context, playerID uuid.UUID) error

	// Get returns the entry for a specific player, or nil if not found.
	Get(ctx context.Context, playerID uuid.UUID) (*entities.ActiveQueueEntry, error)

	// GetAll returns a snapshot of all active queue entries.
	GetAll(ctx context.Context) ([]*entities.ActiveQueueEntry, error)

	// Count returns the number of players currently in the active queue.
	Count(ctx context.Context) (int, error)

	// UpdatePosition updates the position for a specific player.
	// Returns false if the player is not in the store.
	UpdatePosition(ctx context.Context, playerID uuid.UUID, position int) (bool, error)
}
