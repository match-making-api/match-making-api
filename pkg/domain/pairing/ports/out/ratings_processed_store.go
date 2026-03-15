package pairing_out

import (
	"context"

	"github.com/google/uuid"
)

// RatingsProcessedStore tracks which matches have had ratings updated (idempotency).
type RatingsProcessedStore interface {
	// HasProcessed returns true if ratings have already been calculated for this match.
	HasProcessed(ctx context.Context, matchID uuid.UUID) (bool, error)

	// MarkProcessed records that ratings were processed for this match.
	MarkProcessed(ctx context.Context, matchID uuid.UUID) error
}
