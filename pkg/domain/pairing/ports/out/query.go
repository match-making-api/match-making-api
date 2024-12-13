package pairing_out_ports

import (
	pairing_entities "github.com/psavelis/match-making-api/pkg/domain/pairing/entities"
	schedule_entities "github.com/psavelis/match-making-api/pkg/domain/schedules/entities"
)

type PoolReader interface {
	FindPoolBySchedule(s *schedule_entities.Schedule) *pairing_entities.Pool
}
