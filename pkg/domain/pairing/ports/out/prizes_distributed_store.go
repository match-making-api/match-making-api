package pairing_out

import (
	"context"

	"github.com/google/uuid"
)

// PrizesDistributedStore tracks which matches have had prize distribution events published (idempotency).
type PrizesDistributedStore interface {
	// HasDistributed returns true if PrizeDistributed was already published for this match.
	HasDistributed(ctx context.Context, matchID uuid.UUID) (bool, error)

	// MarkDistributed records that prize distribution was published for this match.
	MarkDistributed(ctx context.Context, matchID uuid.UUID) error
}
