package pairing_value_objects

import (
	lobbies_entities "github.com/psavelis/match-making-api/pkg/domain/lobbies/entities"
	schedule_entities "github.com/psavelis/match-making-api/pkg/domain/schedules/entities"
)

type Criteria struct {
	Schedule schedule_entities.Schedule
	Region   lobbies_entities.Region
	Game     lobbies_entities.Game
	// TODO: ideate modes/rank/rating
}
