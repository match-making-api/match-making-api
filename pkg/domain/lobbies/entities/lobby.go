package entities

import (
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
)

type Lobby struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	ClientID uuid.UUID
	Region   entities.Region
}
