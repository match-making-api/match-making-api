package entities

import (
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

type Region struct {
	common.BaseEntity
	ID          uuid.UUID `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Slug        string    `json:"slug" bson:"slug"`
	Description string    `json:"description" bson:"description"`
}
