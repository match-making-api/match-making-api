package schedules_in_ports

import (
	"github.com/google/uuid"
	schedule_entities "github.com/psavelis/match-making-api/pkg/domain/schedules/entities"
)

type PartyScheduleReader interface {
	GetScheduleByPartyID(id uuid.UUID) *schedule_entities.Schedule
}

type PeerScheduleReader interface {
	GetScheduleByPeerID(id uuid.UUID) *schedule_entities.Schedule
}
