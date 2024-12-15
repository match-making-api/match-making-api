package pairing_usecases

import (
	"fmt"

	"github.com/google/uuid"
	// party_entities "github.com/psavelis/match-making-api/pkg/domain/parties/entities"
	pairing_entities "github.com/psavelis/match-making-api/pkg/domain/pairing/entities"
	pairing_out_ports "github.com/psavelis/match-making-api/pkg/domain/pairing/ports/out"
	schedules_in_ports "github.com/psavelis/match-making-api/pkg/domain/schedules/ports/in"
)

type FindPairUseCase struct {
	PartyScheduleReader schedules_in_ports.PartyScheduleReader
	PoolReader          pairing_out_ports.PoolReader
}

type FindPairPayload struct {
	PartyID uuid.UUID
	// PeerID  uuid.UUID

	
}

func (uc *FindPairUseCase) Execute(p FindPairPayload) (*pairing_entities.Pair, error) {
	schedule := uc.PartyScheduleReader.GetScheduleByPartyID(p.PartyID)

	pool := uc.PoolReader.FindPoolBySchedule(schedule) // TODO: by Options (aggregate)

	if pool == nil {
		// pool := // create pool (according to sharding strategy) 
	}




	return nil, 
}
