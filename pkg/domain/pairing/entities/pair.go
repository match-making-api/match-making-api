package pairing_entities

import (
	"github.com/google/uuid"
	party_entities "github.com/psavelis/match-making-api/pkg/domain/parties/entities"
)

type Pair struct {
	Match map[uuid.UUID]*party_entities.Party
}

func NewPair(size int) *Pair {
	return &Pair{
		Match: make(map[uuid.UUID]*party_entities.Party, size),
	}
}
