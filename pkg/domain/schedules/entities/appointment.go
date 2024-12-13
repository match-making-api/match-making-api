package schedule_entities

import (
	"github.com/google/uuid"
	party_entities "github.com/psavelis/match-making-api/pkg/domain/parties/entities"
)

// Single/Manually (DM / Direct Request)

type Appointment struct {
	ID      uuid.UUID
	Parties []party_entities.Party
	Peers   []party_entities.Peer
	// Match
}
