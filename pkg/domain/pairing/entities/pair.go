package entities

import (
	"github.com/google/uuid"
	"github.com/leetgaming/match-making-api/pkg/domain/parties/entities"
)

type Pair struct {
	Match map[uuid.UUID]*entities.Party
}

func NewPair(size int) *Pair {
	return &Pair{
		Match: make(map[uuid.UUID]*entities.Party, size),
	}
}
