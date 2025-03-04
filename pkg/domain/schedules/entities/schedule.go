package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
)

type ScheduleType = int

const (
	Availability ScheduleType = iota
	Constraint
)

type Schedule struct {
	ID      uuid.UUID
	Type    ScheduleType
	Party   *entities.Party
	Peer    *entities.Peer
	Options map[int]DateOption
}

type DateOption struct {
	Months     []time.Month
	Weekdays   []time.Weekday
	Days       []int
	TimeFrames []TimeFrame
}

type TimeFrame struct {
	Start time.Time
	End   time.Time
}
