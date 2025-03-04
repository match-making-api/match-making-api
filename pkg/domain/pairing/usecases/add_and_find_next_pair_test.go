package usecases_test

// import (
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"

// 	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
// 	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
// 	party_entities "github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
// )

// func TestAddAndFindNextPairUseCase_Execute(t *testing.T) {
// 	// Define test cases
// 	type testCase struct {
// 		name              string
// 		partyID           uuid.UUID
// 		criteria          pairing_value_objects.Criteria
// 		scheduleMock      func(m *mocks.MockPartyScheduleReader)
// 		poolReaderMock    func(m *mocks.MockPoolReader)
// 		poolInitiatorMock func(m *mocks.MockPoolInitiator)
// 		pairCreatorMock   func(m *mocks.MockPairCreator)
// 		expectedPair      *pairing_entities.Pair
// 		expectedPool      *pairing_value_objects.Criteria
// 		expectedPosition  int
// 		expectedErr       error
// 	}

// 	testCases := []testCase{
// 		// Test Case 1: Successful Match (Existing Pool)
// 		{
// 			name:     "Successful Match (Existing Pool)",
// 			partyID:  uuid.New(),
// 			criteria: pairing_value_objects.Criteria{PairSize: 2},
// 			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
// 				m.On("GetScheduleByPartyID", mock.AnythingOfType("uuid.UUID")).Return(&pairing_entities.Schedule{ID: uuid.New()})
// 			},
// 			poolReaderMock: func(m *mocks.MockPoolReader) {
// 				matchingPool := &pairing_entities.Pool{
// 					Criteria: pairing_value_objects.Criteria{PairSize: 2},
// 				}
// 				m.On("FindPool", mock.AnythingOfType("pairing_value_objects.Criteria")).Return(matchingPool, nil)
// 			},
// 			pairCreatorMock: func(m *mocks.MockPairCreator) {
// 				party1 := &party_entities.Party{ID: uuid.New()}
// 				party2 := &party_entities.Party{ID: uuid.New()}
// 				expectedPair := &pairing_entities.Pair{Parties: []*party_entities.Party{party1, party2}}
// 				m.On("Execute", mock.AnythingOfType("[]*party_entities.Party")).Return(expectedPair, nil)
// 			},
// 			expectedPair:     &pairing_entities.Pair{Parties: []*party_entities.Party{mock.Anything, mock.Anything}},
// 			expectedPool:     &pairing_value_objects.Criteria{PairSize: 2},
// 			expectedPosition: 2,
// 			expectedErr:      nil,
// 		},
// 		// Test Case 2: No Matching Pool Found (Pool Initiated)
// 		{
// 			name:     "No Matching Pool Found (Pool Initiated)",
// 			partyID:  uuid.New(),
// 			criteria: pairing_value_objects.Criteria{PairSize: 2},
// 			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
// 				m.On("GetScheduleByPartyID", mock.AnythingOfType("uuid.UUID")).Return(&pairing_entities.Schedule{ID: uuid.New()})
// 			},
// 			poolReaderMock: func(m *mocks.MockPoolReader) {
// 				m.On("FindPool", mock.AnythingOfType("pairing_value_objects.Criteria")).Return(nil, nil)
// 			},
// 			poolInitiatorMock: func(m *mocks.MockPoolInitiator) {
// 				pool := &pairing_entities.Pool{ID: uuid.New()}
// 				m.On("Execute", mock.AnythingOfType("pairing_value_objects.Criteria")).Return(pool, nil)
// 			},
// 			pairCreatorMock: func(m *mocks.MockPairCreator) {
// 				t.Fatal("PairCreator shouldn't be called if no matching parties are found in the newly created pool")
// 			},
// 			expectedPair:     nil,
// 			expectedPool:     &pairing_value_objects.Criteria{PairSize: 2},
// 			expectedPosition: 1,
// 			expectedErr:      nil,
// 		},
// 		// Test Case 3: Existing Party in Pool (Wait for another)
// 		{
// 			name:     "Existing Party in Pool (Wait for another)",
// 			partyID:  uuid.New(),
// 			criteria: pairing_value_objects.Criteria{PairSize: 2},
// 			scheduleMock: func(m *mocks.MockPartyScheduleReader) {
// 				m.On("GetScheduleByPartyID", mock.AnythingOfType("uuid.UUID")).Return(&pairing_entities.Schedule{ID: uuid.New()})
// 			},
// 			poolReaderMock: func(m *mocks.MockPoolReader) {
// 				matchingPool := &pairing_entities.Pool{
// 					Parties:  []uuid.UUID{uuid.New()},
// 					Criteria: pairing_value_objects.Criteria{PairSize: 2},
// 				}
// 				m.On("FindPool", mock.AnythingOfType("pairing_value_objects.Criteria")).Return(matchingPool, nil)
// 			},
// 			pairCreatorMock: func(m *mocks.MockPairCreator) {
// 				// PairCreator should not be called as the party is already in the pool
// 			},
// 			expectedPair:     nil,
// 			expectedPool:     &pairing_value_objects.Criteria{PairSize: 2},
// 			expectedPosition: 2,
// 			expectedErr:      nil,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			scheduleMock := new(mocks.MockPartyScheduleReader)
// 			tc.scheduleMock(scheduleMock)

// 			poolReaderMock := new(mocks.MockPoolReader)
// 			tc.poolReaderMock(poolReaderMock)

// 			poolInitiatorMock := new(mocks.MockPoolInitiator)
// 			tc.poolInitiatorMock(poolInitiatorMock)

// 			pairCreatorMock := new(mocks.MockPairCreator)
// 			tc.pairCreatorMock(pairCreatorMock)

// 			uc := AddAndFindNextPairUseCase{
// 				PoolReader:          poolReaderMock,
// 				PoolWriter:          poolReaderMock, // Assuming the same implementation for reading and writing
// 				PartyScheduleReader: scheduleMock,
// 				PoolInitiator:       poolInitiatorMock,
// 				PairCreator:         pairCreatorMock,
// 			}

// 			pair, pool, position, err := uc.Execute(FindPairPayload{
// 				PartyID:  tc.partyID,
// 				Criteria: tc.criteria,
// 			})

// 			assert.Equal(t, tc.expectedPair, pair)
// 			assert.Equal(t, tc.expectedPool, pool.Criteria)
// 			assert.Equal(t, tc.expectedPosition, position)
// 			assert.Equal(t, tc.expectedErr, err)
// 		})
// 	}
// }
