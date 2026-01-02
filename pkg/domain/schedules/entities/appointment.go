package entities

import (
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
)

// Single/Manually (DM / Direct Request)

type Appointment struct {
	ID      uuid.UUID
	Parties []entities.Party
	Peers   []entities.Peer
	// Match
}
