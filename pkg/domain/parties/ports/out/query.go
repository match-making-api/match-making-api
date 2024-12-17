package parties_out

import (
	"github.com/google/uuid"
	party_entities "github.com/psavelis/match-making-api/pkg/domain/parties/entities"
)

type PartyReader interface {
	GetByID(id uuid.UUID) (*party_entities.Party, error)
}

type PeerReader interface {
	GetByID(id uuid.UUID) (*party_entities.Peer, error)
}
