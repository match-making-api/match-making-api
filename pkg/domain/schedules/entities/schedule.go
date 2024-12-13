package schedule_entities

import (
	"time"

	"github.com/google/uuid"
	p "github.com/psavelis/match-making-api/pkg/domain/parties/entities"
)

type ScheduleType = int

const (
	Availability ScheduleType = iota
	Constraint
)

type Schedule struct {
	ID      uuid.UUID
	Type    ScheduleType
	Party   *p.Party
	Peer    *p.Peer
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
