package pairing_out

import (
	"context"

	"github.com/google/uuid"
)

// AnalyticsTrackedStore tracks which matches have had analytics events published (idempotency).
type AnalyticsTrackedStore interface {
	// HasTracked returns true if AnalyticsTracked was already published for this match.
	HasTracked(ctx context.Context, matchID uuid.UUID) (bool, error)

	// MarkTracked records that analytics was published for this match.
	MarkTracked(ctx context.Context, matchID uuid.UUID) error
}
