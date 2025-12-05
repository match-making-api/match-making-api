package value_objects

import (
	"github.com/google/uuid"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
)

type Criteria struct {
	TenantID *uuid.UUID
	ClientID *uuid.UUID
	// Game     *lobbies_entities.Game // TODO: ideate modes/rank/rating_range (game -> tenant + client)
	Schedule *schedule_entities.Schedule
	Region   *game_entities.Region

	PairSize int //Edges    map[int]int
	// MinParties int
	// MaxParties int
}
