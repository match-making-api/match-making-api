package common

import (
	"time"

	"github.com/google/uuid"
)

type BaseEntity struct {
	ID              uuid.UUID              `json:"id" bson:"_id"`
	VisibilityLevel IntendedAudienceKey    `json:"visibility_level" bson:"visibility_level"`
	VisibilityType  VisibilityTypeKey      `json:"visibility_type" bson:"visibility_type"`
	ResourceOwner   ResourceOwner          `json:"resource_owner" bson:"resource_owner"`
	CreatedAt       time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" bson:"updated_at"`
	Includes        map[string]interface{} `json:"includes" bson:"-"`
}

type Entity interface {
	GetID() uuid.UUID
}

func (b BaseEntity) GetID() uuid.UUID {
	return b.ID
}

func NewEntity(resourceOwner ResourceOwner) BaseEntity {
	now := time.Now()
	return BaseEntity{
		ID:              uuid.New(),
		VisibilityLevel: ClientApplicationAudienceIDKey,
		VisibilityType:  CustomVisibilityTypeKey,
		ResourceOwner:   resourceOwner,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
