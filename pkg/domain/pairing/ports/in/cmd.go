package pairing_in

import (
	"github.com/google/uuid"
	pairing_entities "github.com/psavelis/match-making-api/pkg/domain/pairing/entities"
	pairing_value_objects "github.com/psavelis/match-making-api/pkg/domain/pairing/value-objects"
)

type PairCreator interface {
	Execute(pids []uuid.UUID) (*pairing_entities.Pair, error)
}

type PoolInitiator interface {
	Execute(c pairing_value_objects.Criteria) (*pairing_entities.Pool, error)
}

type PartyScheduleMatcher interface {
	Execute(pids []uuid.UUID, qty int, matched []uuid.UUID) (bool, error)
}
