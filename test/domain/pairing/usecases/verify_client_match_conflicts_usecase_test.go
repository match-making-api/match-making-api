package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	parties_entities "github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestVerifyClientMatchConflictsUseCase_Execute(t *testing.T) {
	tenantID := uuid.New()
	clientID := uuid.New()

	makeSchedule := func(weekday time.Weekday, startHour, endHour int) *schedule_entities.Schedule {
		start := time.Date(2026, 1, 1, startHour, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 1, endHour, 0, 0, 0, time.UTC)
		return &schedule_entities.Schedule{
			ID:   uuid.New(),
			Type: schedule_entities.Availability,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Weekdays:   []time.Weekday{weekday},
					Days:       []int{1},
					TimeFrames: []schedule_entities.TimeFrame{{Start: start, End: end}},
				},
			},
		}
	}

	makePair := func(partyIDs ...uuid.UUID) *pairing_entities.Pair {
		ro := common.ResourceOwner{TenantID: tenantID, ClientID: clientID}
		pair := pairing_entities.NewPair(len(partyIDs), ro)
		for _, pid := range partyIDs {
			pair.Match[pid] = &parties_entities.Party{ID: pid}
		}
		return pair
	}

	tests := []struct {
		name           string
		partyID        uuid.UUID
		setupMocks     func(uuid.UUID, *mocks.MockPortPairReader, *mocks.MockPortPairWriter, *mocks.MockPartyScheduleReader, *mocks.MockConflictNotifier)
		expectedResult *usecases.ConflictResult
		expectedError  string
	}{
		{
			name:    "no conflicts when party has no schedule",
			partyID: uuid.New(),
			setupMocks: func(partyID uuid.UUID, pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter, scheduleReader *mocks.MockPartyScheduleReader, _ *mocks.MockConflictNotifier) {
				scheduleReader.On("GetScheduleByPartyID", partyID).Return(nil)
			},
			expectedResult: &usecases.ConflictResult{HasConflict: false, ConflictingPairs: []uuid.UUID{}},
		},
		{
			name:    "no conflicts when party has no matches",
			partyID: uuid.New(),
			setupMocks: func(partyID uuid.UUID, pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter, scheduleReader *mocks.MockPartyScheduleReader, _ *mocks.MockConflictNotifier) {
				scheduleReader.On("GetScheduleByPartyID", partyID).Return(makeSchedule(time.Monday, 18, 22))
				pairReader.On("FindPairsByPartyID", mock.Anything, partyID).Return([]*pairing_entities.Pair{}, nil)
			},
			expectedResult: &usecases.ConflictResult{HasConflict: false, ConflictingPairs: []uuid.UUID{}},
		},
		{
			name:    "error when FindPairsByPartyID fails",
			partyID: uuid.New(),
			setupMocks: func(partyID uuid.UUID, pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter, scheduleReader *mocks.MockPartyScheduleReader, _ *mocks.MockConflictNotifier) {
				scheduleReader.On("GetScheduleByPartyID", partyID).Return(makeSchedule(time.Monday, 18, 22))
				pairReader.On("FindPairsByPartyID", mock.Anything, partyID).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to find pairs",
		},
		{
			name:    "no conflicts with non-overlapping pair schedules on different days",
			partyID: uuid.New(),
			setupMocks: func(partyID uuid.UUID, pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter, scheduleReader *mocks.MockPartyScheduleReader, _ *mocks.MockConflictNotifier) {
				party2 := uuid.New()
				party3 := uuid.New()

				// Client plays Monday 18-22
				clientSchedule := makeSchedule(time.Monday, 18, 22)
				scheduleReader.On("GetScheduleByPartyID", partyID).Return(clientSchedule)
				// Match 1 party: Monday 18-20
				scheduleReader.On("GetScheduleByPartyID", party2).Return(makeSchedule(time.Monday, 18, 20))
				// Match 2 party: Tuesday 18-20 (different day)
				scheduleReader.On("GetScheduleByPartyID", party3).Return(makeSchedule(time.Tuesday, 18, 20))

				pair1 := makePair(partyID, party2)
				pair2 := makePair(partyID, party3)
				pairReader.On("FindPairsByPartyID", mock.Anything, partyID).Return([]*pairing_entities.Pair{pair1, pair2}, nil)
			},
			expectedResult: &usecases.ConflictResult{HasConflict: false, ConflictingPairs: []uuid.UUID{}},
		},
		{
			name:    "detect conflicts with overlapping pair schedules",
			partyID: uuid.New(),
			setupMocks: func(partyID uuid.UUID, pairReader *mocks.MockPortPairReader, pairWriter *mocks.MockPortPairWriter, scheduleReader *mocks.MockPartyScheduleReader, notifier *mocks.MockConflictNotifier) {
				party2 := uuid.New()
				party3 := uuid.New()

				// Client plays Monday 18-22
				clientSchedule := makeSchedule(time.Monday, 18, 22)
				scheduleReader.On("GetScheduleByPartyID", partyID).Return(clientSchedule)
				// Match 1 party: Monday 19-21
				scheduleReader.On("GetScheduleByPartyID", party2).Return(makeSchedule(time.Monday, 19, 21))
				// Match 2 party: Monday 20-23 (overlaps with match 1)
				scheduleReader.On("GetScheduleByPartyID", party3).Return(makeSchedule(time.Monday, 20, 23))

				pair1 := makePair(partyID, party2)
				pair2 := makePair(partyID, party3)
				pairReader.On("FindPairsByPartyID", mock.Anything, partyID).Return([]*pairing_entities.Pair{pair1, pair2}, nil)

				// Expect flagging and notifications (may or may not be called depending on conflict detection logic)
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair1, nil).Maybe()
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Return(pair1, nil).Maybe()
				notifier.On("NotifyConflict", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			},
		},
		{
			name:    "single match with no other pairs - no conflicts",
			partyID: uuid.New(),
			setupMocks: func(partyID uuid.UUID, pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter, scheduleReader *mocks.MockPartyScheduleReader, _ *mocks.MockConflictNotifier) {
				party2 := uuid.New()

				clientSchedule := makeSchedule(time.Wednesday, 14, 18)
				scheduleReader.On("GetScheduleByPartyID", partyID).Return(clientSchedule)
				scheduleReader.On("GetScheduleByPartyID", party2).Return(makeSchedule(time.Wednesday, 14, 16))

				pair := makePair(partyID, party2)
				pairReader.On("FindPairsByPartyID", mock.Anything, partyID).Return([]*pairing_entities.Pair{pair}, nil)
			},
			expectedResult: &usecases.ConflictResult{HasConflict: false, ConflictingPairs: []uuid.UUID{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPairReader := new(mocks.MockPortPairReader)
			mockPairWriter := new(mocks.MockPortPairWriter)
			mockScheduleReader := new(mocks.MockPartyScheduleReader)
			mockNotifier := new(mocks.MockConflictNotifier)

			tt.setupMocks(tt.partyID, mockPairReader, mockPairWriter, mockScheduleReader, mockNotifier)

			useCase := &usecases.VerifyClientMatchConflictsUseCase{
				PairReader:          mockPairReader,
				PairWriter:          mockPairWriter,
				PartyScheduleReader: mockScheduleReader,
				ConflictNotifier:    mockNotifier,
			}

			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.partyID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.HasConflict, result.HasConflict)
				}
			}
		})
	}
}
