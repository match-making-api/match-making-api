package pairing_usecases

import (
	"fmt"

	"github.com/google/uuid"
	pairing_entities "github.com/psavelis/match-making-api/pkg/domain/pairing/entities"
	pairing_in "github.com/psavelis/match-making-api/pkg/domain/pairing/ports/in"
	pairing_out "github.com/psavelis/match-making-api/pkg/domain/pairing/ports/out"
	pairing_value_objects "github.com/psavelis/match-making-api/pkg/domain/pairing/value-objects"
	schedules_in_ports "github.com/psavelis/match-making-api/pkg/domain/schedules/ports/in"
)

type AddAndFindNextPairUseCase struct {
	PoolReader pairing_out.PoolReader
	PoolWriter pairing_out.PoolWriter

	PartyScheduleReader schedules_in_ports.PartyScheduleReader
	PoolInitiator       pairing_in.PoolInitiator
	PairCreator         pairing_in.PairCreator
	ScheduleMatcher     pairing_in.PartyScheduleMatcher
}

type FindPairPayload struct {
	PartyID  uuid.UUID // (always create a party, even if alone, easier to add someone to it if the user decides so in the middle of match making)
	Criteria pairing_value_objects.Criteria
}

func (uc *AddAndFindNextPairUseCase) Execute(p FindPairPayload) (*pairing_entities.Pair, *pairing_entities.Pool, int, error) {
	schedule := uc.PartyScheduleReader.GetScheduleByPartyID(p.PartyID)

	p.Criteria.Schedule = schedule

	var pool *pairing_entities.Pool
	var err error

	pool, _ = uc.PoolReader.FindPool(&p.Criteria)

	if pool == nil {
		pool, err = uc.PoolInitiator.Execute(p.Criteria)

		if err != nil {
			return nil, nil, -1, fmt.Errorf("AddAndFindNextPairUseCase.Execute: unable to FIND pair. Cannot create pool with Criteria %v, due to %v", p.Criteria, err)
		}
	}

	position := pool.Join(p.PartyID) // ADD: party Or peer. (Idempotent => wont dup if already enqueued)
	uc.PoolWriter.Save(pool)

	parties := pool.Peek(p.Criteria.PairSize) // FIND: equiv: pool.Dequeue(s, q)

	var pair *pairing_entities.Pair

	// if succesfuly dequeued
	if len(parties) > 0 {
		pair, err = uc.PairCreator.Execute(parties)
		if err != nil {
			return nil, nil, position, fmt.Errorf("AddAndFindNextPairUseCase.Execute: unable to CREATE pair. Cannot create pair for parties %v, due to %v", parties, err)
		}

		pool, err = uc.PoolWriter.Save(pool)

		if err != nil {
			return nil, nil, position, fmt.Errorf("AddAndFindNextPairUseCase.Execute: unable to UPDATE pool for pair %v. Cannot update pool for parties %v, due to %v", pair, parties, err)
		}
	}

	return pair, pool, position, nil // send msg with position etc?
}
