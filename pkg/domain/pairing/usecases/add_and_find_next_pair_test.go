package usecases_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
	parties_entities "github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestAddAndFindNextPairUseCase_Execute(t *testing.T) {
	// Define test cases
	type testCase struct {
		name              string
		partyID           uuid.UUID
		criteria          pairing_value_objects.Criteria
		scheduleMock      func(m *mocks.MockPartyScheduleReader)
		poolReaderMock    func(m *mocks.MockPoolReader)
		poolInitiatorMock func(m *mocks.MockPoolInitiator)
		pairCreatorMock   func(m *mocks.MockPairCreator)
		poolWriterMock    func(m *mocks.MockPoolWriter)
		expectedPair      *pairing_entities.Pair
		expectedPool      *pairing_entities.Pool
		expectedPosition  int
		expectedErr       error
	}

	testCases := []testCase{
		{
			name:     "Successful Match (Existing Pool)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil)
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 2),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 2,
			},
			expectedPosition: 2,
			expectedErr:      nil,
		},
		{
			name:     "No Matching Pool Found (Pool Initiated)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				m.On("FindPool", mock.Anything).Return(nil, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("Execute", mock.Anything).Return(pool, nil)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called when pool is empty
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil)
			},
			expectedPair:     nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "Pool Initiator Fails",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				m.On("FindPool", mock.Anything).Return(nil, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				m.On("Execute", mock.Anything).Return(nil, assert.AnError)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// Should not be called
			},
			expectedPair:     nil,
			expectedPool:     nil,
			expectedPosition: -1,
			expectedErr:      assert.AnError,
		},
		{
			name:     "Pair Creator Fails",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				m.On("Execute", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(nil, assert.AnError)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// Pool is saved after joining, but not after pair creation failure
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New(), uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair:     nil,
			expectedPool:     nil,
			expectedPosition: 2,
			expectedErr:      assert.AnError,
		},
		{
			name:     "Pool Writer Fails After Pair Creation",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				m.On("Save", mock.Anything).Return(nil, assert.AnError)
			},
			expectedPair:     nil,
			expectedPool:     nil,
			expectedPosition: 2,
			expectedErr:      assert.AnError,
		},
		{
			name:     "Schedule Reader Fails",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				m.On("GetScheduleByPartyID", mock.Anything).Return(nil)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// FindPool is still called even with nil schedule
				m.On("FindPool", mock.Anything).Return(nil, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("Execute", mock.Anything).Return(pool, nil)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair:     nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "Pool Reader Fails",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				m.On("FindPool", mock.Anything).Return(nil, assert.AnError)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("Execute", mock.Anything).Return(pool, nil)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil)
			},
			expectedPair:     nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "First Pool Save Fails But Continues",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save fails but doesn't stop execution
				m.On("Save", mock.Anything).Return(nil, assert.AnError).Once()
				// Second save after pair creation also fails
				m.On("Save", mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectedPair:     nil,
			expectedPool:     nil,
			expectedPosition: 2,
			expectedErr:      assert.AnError, // Error from second save
		},
		{
			name:     "Idempotent Join - Party Already in Pool",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				partyID := uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{partyID}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				poolAfterJoin.Parties = []uuid.UUID{uuid.New(), uuid.New()}
				poolAfterJoin.PartySize = 2

				// Second save after pair creation
				poolAfterPair := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				poolAfterPair.Parties = []uuid.UUID{}
				poolAfterPair.PartySize = 2

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 2
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 0
				})).Return(poolAfterPair, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 2),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 2,
			},
			expectedPosition: 2,
			expectedErr:      nil,
		},
		{
			name:     "Real Scenario: First Player Joins Empty Pool",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				m.On("FindPool", mock.Anything).Return(nil, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("Execute", mock.Anything).Return(pool, nil)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called - only 1 party
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair:     nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "Real Scenario: Second Player Joins, Creates Pair",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				partyID1 := uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{partyID1}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 2
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				poolAfterJoin.Parties = []uuid.UUID{uuid.New(), uuid.New()}
				poolAfterJoin.PartySize = 2

				// Second save after pair creation (pool should be empty)
				poolAfterPair := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				poolAfterPair.Parties = []uuid.UUID{}
				poolAfterPair.PartySize = 2

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 2
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 0
				})).Return(poolAfterPair, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 2),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 2,
			},
			expectedPosition: 2,
			expectedErr:      nil,
		},
		{
			name:     "Real Scenario: Third Player Joins After Pair Creation",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{} // Empty after previous pair creation
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called - only 1 party
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair:     nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "Pool With Different Pair Size Criteria",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 3}, // Looking for groups of 3
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				m.On("FindPool", mock.Anything).Return(nil, nil) // No pool with pair size 3
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 3})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 3
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("Execute", mock.Anything).Return(pool, nil)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called - only 1 party, need 3
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 3})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 3
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair:     nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 3,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "Pool Peek Returns Empty Despite Having Parties",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 3}, // Need 3 parties but only have 2
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 3})
				// Pool has 2 parties, but we need 3 for a match
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 3
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called - not enough parties for pair (only 2, need 3)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 3})
				pool.Parties = []uuid.UUID{uuid.New(), uuid.New()}
				pool.PartySize = 3
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair:     nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New(), uuid.New()},
				PartySize: 3,
			},
			expectedPosition: 2,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Multiple Players Already Waiting (4 Players, Pair Size 2)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// Pool already has 3 players waiting
				party1, party2, party3 := uuid.New(), uuid.New(), uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = []uuid.UUID{party1, party2, party3}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 2
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join (now 4 players)
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				poolAfterJoin.Parties = []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
				poolAfterJoin.PartySize = 2

				// Second save after pair creation (2 players removed, 2 remain)
				poolAfterPair := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				poolAfterPair.Parties = []uuid.UUID{uuid.New(), uuid.New()}
				poolAfterPair.PartySize = 2

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 4
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 2
				})).Return(poolAfterPair, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 2),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New(), uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 4,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Tournament Matchmaking (3-Player Teams)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 3},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// Pool has 2 players waiting for 3-player teams
				party1, party2 := uuid.New(), uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 3})
				pool.Parties = []uuid.UUID{party1, party2}
				pool.PartySize = 3
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(3, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 3
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join (now 3 players)
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 3})
				poolAfterJoin.Parties = []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
				poolAfterJoin.PartySize = 3

				// Second save after team creation (pool emptied)
				poolAfterTeam := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 3})
				poolAfterTeam.Parties = []uuid.UUID{}
				poolAfterTeam.PartySize = 3

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 3
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 0
				})).Return(poolAfterTeam, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 3),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 3,
			},
			expectedPosition: 3,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Regional Matchmaking (North America)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
				Region:   &game_entities.Region{Name: "North America"},
			},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// No existing pool for North America region
				m.On("FindPool", mock.Anything).Return(nil, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					Region:   &game_entities.Region{Name: "North America"},
				})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("Execute", mock.Anything).Return(pool, nil)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called - only 1 player
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					Region:   &game_entities.Region{Name: "North America"},
				})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair: nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Tenant-Specific Matchmaking",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
				TenantID: &[]uuid.UUID{uuid.New()}[0],
			},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// No existing tenant-specific pool
				m.On("FindPool", mock.Anything).Return(nil, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				tenantID := uuid.New()
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					TenantID: &tenantID,
				})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("Execute", mock.Anything).Return(pool, nil)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called - only 1 player
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				tenantID := uuid.New()
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					TenantID: &tenantID,
				})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair: nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: High-Concurrency Peak Hours (Many Players Joining)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{PairSize: 2},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// Simulate a busy pool with many players already waiting
				parties := make([]uuid.UUID, 10)
				for i := range parties {
					parties[i] = uuid.New()
				}
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				pool.Parties = parties
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 2
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join (now 11 players)
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				poolAfterJoin.Parties = make([]uuid.UUID, 11)
				for i := range poolAfterJoin.Parties {
					poolAfterJoin.Parties[i] = uuid.New()
				}
				poolAfterJoin.PartySize = 2

				// Second save after pair creation (9 players remain)
				poolAfterPair := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{PairSize: 2})
				poolAfterPair.Parties = make([]uuid.UUID, 9)
				for i := range poolAfterPair.Parties {
					poolAfterPair.Parties[i] = uuid.New()
				}
				poolAfterPair.PartySize = 2

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 11
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 9
				})).Return(poolAfterPair, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 2),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   make([]uuid.UUID, 9),
				PartySize: 2,
			},
			expectedPosition: 11,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Client-Specific Matchmaking",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
				ClientID: &[]uuid.UUID{uuid.New()}[0],
			},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// Client joins existing client-specific pool
				party1 := uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				clientID := uuid.New()
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					ClientID: &clientID,
				})
				pool.Parties = []uuid.UUID{party1}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 2
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				clientID := uuid.New()
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					ClientID: &clientID,
				})
				poolAfterJoin.Parties = []uuid.UUID{uuid.New(), uuid.New()}
				poolAfterJoin.PartySize = 2

				// Second save after pair creation
				poolAfterPair := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					ClientID: &clientID,
				})
				poolAfterPair.Parties = []uuid.UUID{}
				poolAfterPair.PartySize = 2

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 2
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 0
				})).Return(poolAfterPair, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 2),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 2,
			},
			expectedPosition: 2,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Competitive CS2 Matchmaking (Ranked Mode)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{
				PairSize: 5,
				GameID:   &[]uuid.UUID{uuid.New()}[0],
				GameModeID: &[]uuid.UUID{uuid.New()}[0],
				Region:   &game_entities.Region{Name: "North America"},
				SkillRange: &pairing_value_objects.SkillRange{
					MinMMR: 1400,
					MaxMMR: 1600,
				},
				MaxPing:           50,
				AllowCrossPlatform: true,
				Tier:              "premium",
				PriorityBoost:     false,
			},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// Pool has 4 players waiting for competitive CS2
				party1, party2, party3, party4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				cs2GameID := uuid.New()
				competitiveModeID := uuid.New()
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 5,
					GameID:   &cs2GameID,
					GameModeID: &competitiveModeID,
					Region:   &game_entities.Region{Name: "North America"},
					SkillRange: &pairing_value_objects.SkillRange{
						MinMMR: 1400,
						MaxMMR: 1600,
					},
				})
				pool.Parties = []uuid.UUID{party1, party2, party3, party4}
				pool.PartySize = 5
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(5, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 5
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join (now 5 players)
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				cs2GameID := uuid.New()
				competitiveModeID := uuid.New()
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 5,
					GameID:   &cs2GameID,
					GameModeID: &competitiveModeID,
				})
				poolAfterJoin.Parties = make([]uuid.UUID, 5)
				for i := range poolAfterJoin.Parties {
					poolAfterJoin.Parties[i] = uuid.New()
				}
				poolAfterJoin.PartySize = 5

				// Second save after team creation (pool emptied)
				poolAfterTeam := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 5,
					GameID:   &cs2GameID,
					GameModeID: &competitiveModeID,
				})
				poolAfterTeam.Parties = []uuid.UUID{}
				poolAfterTeam.PartySize = 5

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 5
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 0
				})).Return(poolAfterTeam, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 5),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 5,
			},
			expectedPosition: 5,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Casual Valorant Matchmaking (Quick Play)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{
				PairSize: 5,
				GameID:   &[]uuid.UUID{uuid.New()}[0],
				GameModeID: &[]uuid.UUID{uuid.New()}[0],
				Region:   &game_entities.Region{Name: "Europe"},
				SkillRange: &pairing_value_objects.SkillRange{
					MinMMR: 800,
					MaxMMR: 1200,
				},
				MaxPing:           60,
				AllowCrossPlatform: false,
				Tier:              "free",
				PriorityBoost:     false,
			},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// Pool has 4 players waiting for casual Valorant
				party1, party2, party3, party4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				valorantGameID := uuid.New()
				casualModeID := uuid.New()
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 5,
					GameID:   &valorantGameID,
					GameModeID: &casualModeID,
					Region:   &game_entities.Region{Name: "Europe"},
					SkillRange: &pairing_value_objects.SkillRange{
						MinMMR: 800,
						MaxMMR: 1200,
					},
				})
				pool.Parties = []uuid.UUID{party1, party2, party3, party4}
				pool.PartySize = 5
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(5, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 5
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				valorantGameID := uuid.New()
				casualModeID := uuid.New()
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 5,
					GameID:   &valorantGameID,
					GameModeID: &casualModeID,
				})
				poolAfterJoin.Parties = make([]uuid.UUID, 5)
				for i := range poolAfterJoin.Parties {
					poolAfterJoin.Parties[i] = uuid.New()
				}
				poolAfterJoin.PartySize = 5

				// Second save after team creation
				poolAfterTeam := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 5,
					GameID:   &valorantGameID,
					GameModeID: &casualModeID,
				})
				poolAfterTeam.Parties = []uuid.UUID{}
				poolAfterTeam.PartySize = 5

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 5
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 0
				})).Return(poolAfterTeam, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 5),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 5,
			},
			expectedPosition: 5,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Premium Tier Priority Matchmaking",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
				GameID:   &[]uuid.UUID{uuid.New()}[0],
				GameModeID: &[]uuid.UUID{uuid.New()}[0],
				Region:   &game_entities.Region{Name: "Asia"},
				SkillRange: &pairing_value_objects.SkillRange{
					MinMMR: 1800,
					MaxMMR: 2200,
				},
				MaxPing:           30,
				AllowCrossPlatform: true,
				Tier:              "premium",
				PriorityBoost:     true,
			},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// No existing premium pool for high-skill ranked
				m.On("FindPool", mock.Anything).Return(nil, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					GameID:   &[]uuid.UUID{uuid.New()}[0],
					GameModeID: &[]uuid.UUID{uuid.New()}[0],
					Region:   &game_entities.Region{Name: "Asia"},
					SkillRange: &pairing_value_objects.SkillRange{
						MinMMR: 1800,
						MaxMMR: 2200,
					},
					Tier:          "premium",
					PriorityBoost: true,
				})
				pool.Parties = []uuid.UUID{}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("Execute", mock.Anything).Return(pool, nil)
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				// Should not be called - only 1 player
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					GameID:   &[]uuid.UUID{uuid.New()}[0],
					GameModeID: &[]uuid.UUID{uuid.New()}[0],
					Tier:     "premium",
				})
				pool.Parties = []uuid.UUID{uuid.New()}
				pool.PartySize = 2
				m.On("Save", mock.Anything).Return(pool, nil).Once()
			},
			expectedPair: nil,
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{uuid.New()},
				PartySize: 2,
			},
			expectedPosition: 1,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Map-Specific Matchmaking (Dust2 Only)",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
				GameID:   &[]uuid.UUID{uuid.New()}[0],
				GameModeID: &[]uuid.UUID{uuid.New()}[0],
				Region:   &game_entities.Region{Name: "North America"},
				MapPreferences: []string{"dust2"},
				SkillRange: &pairing_value_objects.SkillRange{
					MinMMR: 1000,
					MaxMMR: 1400,
				},
				MaxPing:           40,
				AllowCrossPlatform: true,
				Tier:              "free",
				PriorityBoost:     false,
			},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Availability,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// Pool has 1 player waiting for Dust2 matches
				party1 := uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					GameID:   &[]uuid.UUID{uuid.New()}[0],
					GameModeID: &[]uuid.UUID{uuid.New()}[0],
					MapPreferences: []string{"dust2"},
					SkillRange: &pairing_value_objects.SkillRange{
						MinMMR: 1000,
						MaxMMR: 1400,
					},
				})
				pool.Parties = []uuid.UUID{party1}
				pool.PartySize = 2
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(2, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 2
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					GameID:   &[]uuid.UUID{uuid.New()}[0],
					GameModeID: &[]uuid.UUID{uuid.New()}[0],
					MapPreferences: []string{"dust2"},
				})
				poolAfterJoin.Parties = []uuid.UUID{uuid.New(), uuid.New()}
				poolAfterJoin.PartySize = 2

				// Second save after pair creation
				poolAfterPair := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 2,
					GameID:   &[]uuid.UUID{uuid.New()}[0],
					GameModeID: &[]uuid.UUID{uuid.New()}[0],
					MapPreferences: []string{"dust2"},
				})
				poolAfterPair.Parties = []uuid.UUID{}
				poolAfterPair.PartySize = 2

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 2
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 0
				})).Return(poolAfterPair, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 2),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 2,
			},
			expectedPosition: 2,
			expectedErr:      nil,
		},
		{
			name:     "Real-World: Low Ping Tournament Matchmaking",
			partyID:  uuid.New(),
			criteria: pairing_value_objects.Criteria{
				PairSize: 4,
				GameID:   &[]uuid.UUID{uuid.New()}[0],
				GameModeID: &[]uuid.UUID{uuid.New()}[0],
				Region:   &game_entities.Region{Name: "North America"},
				SkillRange: &pairing_value_objects.SkillRange{
					MinMMR: 1600,
					MaxMMR: 2000,
				},
				MaxPing:           20, // Very low ping requirement
				AllowCrossPlatform: false,
				Tier:              "pro",
				PriorityBoost:     true,
			},
			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
				schedule := &schedule_entities.Schedule{
					ID:   uuid.New(),
					Type: schedule_entities.Constraint,
				}
				m.On("GetScheduleByPartyID", mock.Anything).Return(schedule)
			},
			poolReaderMock: func(m *mocks.MockPoolReader) {
				// Pool has 3 players waiting for tournament
				party1, party2, party3 := uuid.New(), uuid.New(), uuid.New()
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				pool := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 4,
					GameID:   &[]uuid.UUID{uuid.New()}[0],
					GameModeID: &[]uuid.UUID{uuid.New()}[0],
					MaxPing:  20,
					Tier:     "pro",
				})
				pool.Parties = []uuid.UUID{party1, party2, party3}
				pool.PartySize = 4
				pool.CreatedAt = time.Now()
				pool.UpdatedAt = time.Now()
				m.On("FindPool", mock.Anything).Return(pool, nil)
			},
			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
				// Should not be called
			},
			pairCreatorMock: func(m *mocks.MockPairCreator) {
				resourceOwner := common.ResourceOwner{
					TenantID: uuid.New(),
					ClientID: uuid.New(),
					UserID:   uuid.New(),
				}
				pair := pairing_entities.NewPair(4, resourceOwner)
				m.On("Execute", mock.Anything, mock.MatchedBy(func(parties []uuid.UUID) bool {
					return len(parties) == 4
				})).Return(pair, nil)
			},
			poolWriterMock: func(m *mocks.MockPoolWriter) {
				// First save after join
				mutex := &sync.Mutex{}
				cond := sync.NewCond(mutex)
				poolAfterJoin := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 4,
					GameID:   &[]uuid.UUID{uuid.New()}[0],
					GameModeID: &[]uuid.UUID{uuid.New()}[0],
					MaxPing:  20,
					Tier:     "pro",
				})
				poolAfterJoin.Parties = make([]uuid.UUID, 4)
				for i := range poolAfterJoin.Parties {
					poolAfterJoin.Parties[i] = uuid.New()
				}
				poolAfterJoin.PartySize = 4

				// Second save after team creation
				poolAfterTeam := pairing_entities.NewPool(mutex, cond, pairing_value_objects.Criteria{
					PairSize: 4,
					GameID:   &[]uuid.UUID{uuid.New()}[0],
					GameModeID: &[]uuid.UUID{uuid.New()}[0],
					MaxPing:  20,
					Tier:     "pro",
				})
				poolAfterTeam.Parties = []uuid.UUID{}
				poolAfterTeam.PartySize = 4

				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 4
				})).Return(poolAfterJoin, nil).Once()
				m.On("Save", mock.MatchedBy(func(p *pairing_entities.Pool) bool {
					return len(p.Parties) == 0
				})).Return(poolAfterTeam, nil).Once()
			},
			expectedPair: &pairing_entities.Pair{
				Match:          make(map[uuid.UUID]*parties_entities.Party, 4),
				ConflictStatus: pairing_entities.ConflictStatusNone,
			},
			expectedPool: &pairing_entities.Pool{
				Parties:   []uuid.UUID{},
				PartySize: 4,
			},
			expectedPosition: 4,
			expectedErr:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheduleMock := &mocks.MockPartyScheduleReader{}
			tc.scheduleMock(scheduleMock)

			poolReaderMock := &mocks.MockPoolReader{}
			tc.poolReaderMock(poolReaderMock)

			poolInitiatorMock := &mocks.MockPoolInitiator{}
			tc.poolInitiatorMock(poolInitiatorMock)

			pairCreatorMock := &mocks.MockPairCreator{}
			tc.pairCreatorMock(pairCreatorMock)

			poolWriterMock := &mocks.MockPoolWriter{}
			tc.poolWriterMock(poolWriterMock)

			uc := usecases.AddAndFindNextPairUseCase{
				PoolReader:         poolReaderMock,
				PoolWriter:         poolWriterMock,
				PartyScheduleReader: scheduleMock,
				PoolInitiator:      poolInitiatorMock,
				PairCreator:        pairCreatorMock,
			}

			pair, pool, position, err := uc.Execute(usecases.FindPairPayload{
				PartyID:  tc.partyID,
				Criteria: tc.criteria,
			})

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedPair != nil {
				assert.NotNil(t, pair)
				assert.Equal(t, tc.expectedPair.ConflictStatus, pair.ConflictStatus)
			} else {
				assert.Nil(t, pair)
			}

			if tc.expectedPool != nil {
				assert.NotNil(t, pool)
				assert.Equal(t, tc.expectedPool.PartySize, pool.PartySize)
			} else {
				assert.Nil(t, pool)
			}

			assert.Equal(t, tc.expectedPosition, position)
		})
	}
}
