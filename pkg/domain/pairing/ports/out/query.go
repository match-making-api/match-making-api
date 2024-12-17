package pairing_out

import (
	"github.com/google/uuid"
	pairing_entities "github.com/psavelis/match-making-api/pkg/domain/pairing/entities"
	pairing_value_objects "github.com/psavelis/match-making-api/pkg/domain/pairing/value-objects"
)

type PoolReader interface {
	FindPool(c *pairing_value_objects.Criteria) (*pairing_entities.Pool, error)
}

type PairReader interface {
	GetByID(id uuid.UUID) (*pairing_entities.Pair, error)
}
