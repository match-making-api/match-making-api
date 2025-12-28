package pairing_out

import (
	"context"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
)

type PoolWriter interface {
	Save(p *pairing_entities.Pool) (*pairing_entities.Pool, error)
}

type PairWriter interface {
	Save(p *pairing_entities.Pair) (*pairing_entities.Pair, error)
}

type PairReader interface {
	FindPairsByPartyID(ctx context.Context, partyID uuid.UUID) ([]*pairing_entities.Pair, error)
	GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Pair, error)
}

type InvitationWriter interface {
	Save(ctx context.Context, invitation *pairing_entities.Invitation) (*pairing_entities.Invitation, error)
}

type InvitationReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Invitation, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*pairing_entities.Invitation, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]*pairing_entities.Invitation, error)
}

type PoolReader interface {
	FindPool(criteria *pairing_value_objects.Criteria) (*pairing_entities.Pool, error)
}
