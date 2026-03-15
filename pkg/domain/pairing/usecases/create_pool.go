package usecases

import (
	"fmt"
	"sync"

	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
)

type CreatePoolUseCase struct {
	PoolWriter pairing_out.PoolWriter
}

type CreatePoolPayload struct {
	Criteria *pairing_value_objects.Criteria
}

// Execute creates a new pool for the given matchmaking criteria.
// It implements the PoolInitiator port interface.
func (uc *CreatePoolUseCase) Execute(c pairing_value_objects.Criteria) (*pairing_entities.Pool, error) {
	if c.PairSize <= 0 {
		return nil, fmt.Errorf("CreatePoolUseCase.Execute: pair size must be greater than 0, got %d", c.PairSize)
	}

	mutex := &sync.Mutex{}
	cond := sync.NewCond(mutex)
	pool := pairing_entities.NewPool(mutex, cond, c)

	pool.PartySize = uint8(c.PairSize)

	if uc.PoolWriter != nil {
		savedPool, err := uc.PoolWriter.Save(pool)
		if err != nil {
			return nil, fmt.Errorf("CreatePoolUseCase.Execute: failed to save pool: %w", err)
		}
		return savedPool, nil
	}

	return pool, nil
}
