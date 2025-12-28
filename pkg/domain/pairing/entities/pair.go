package entities

import (
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
)

type ConflictStatus int

const (
	ConflictStatusNone ConflictStatus = iota
	ConflictStatusFlagged
	ConflictStatusResolved
)

type Pair struct {
	common.BaseEntity
	Match          map[uuid.UUID]*entities.Party `json:"match" bson:"match"`
	ConflictStatus ConflictStatus                `json:"conflict_status" bson:"conflict_status"`
	ConflictReason string                        `json:"conflict_reason,omitempty" bson:"conflict_reason,omitempty"`
}

func NewPair(size int, resourceOwner common.ResourceOwner) *Pair {
	return &Pair{
		BaseEntity:     common.NewEntity(resourceOwner),
		Match:          make(map[uuid.UUID]*entities.Party, size),
		ConflictStatus: ConflictStatusNone,
	}
}
