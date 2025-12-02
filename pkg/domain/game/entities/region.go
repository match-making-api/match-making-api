package entities

import (
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// Region represents a geographical region where a game can be played.
type Region struct {
	common.BaseEntity
	Name        string `json:"name" bson:"name"`               // Unique name for the region
	Slug        string `json:"slug" bson:"slug"`               // Slug for the region (e.g., "eu-west-1")
	Description string `json:"description" bson:"description"` // Description of the region
}
