package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
	schedules_in_ports "github.com/leet-gaming/match-making-api/pkg/domain/schedules/ports/in"
)

// MockPartyScheduleReader is a mock implementation of schedules_in_ports.PartyScheduleReader
type MockPartyScheduleReader struct {
	mock.Mock
	schedules map[uuid.UUID]*schedule_entities.Schedule
}

// Ensure MockPartyScheduleReader implements schedules_in_ports.PartyScheduleReader
var _ schedules_in_ports.PartyScheduleReader = (*MockPartyScheduleReader)(nil)

// NewMockPartyScheduleReader creates a new mock with the given schedules
func NewMockPartyScheduleReader(schedules map[uuid.UUID]*schedule_entities.Schedule) *MockPartyScheduleReader {
	return &MockPartyScheduleReader{
		schedules: schedules,
	}
}

// GetScheduleByPartyID returns the schedule for the given party ID
func (m *MockPartyScheduleReader) GetScheduleByPartyID(id uuid.UUID) *schedule_entities.Schedule {
	// Check if there are any expectations set
	if len(m.ExpectedCalls) > 0 {
		args := m.Called(id)
		if args.Get(0) != nil {
			return args.Get(0).(*schedule_entities.Schedule)
		}
	}
	
	// Fallback to internal map if no mock expectation is set or expectation returned nil
	if m.schedules != nil {
		return m.schedules[id]
	}
	return nil
}
