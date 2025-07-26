package entities

import (
	"github.com/gofrs/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

type GameMode struct {
	common.BaseEntity
	ID          uuid.UUID `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
}
