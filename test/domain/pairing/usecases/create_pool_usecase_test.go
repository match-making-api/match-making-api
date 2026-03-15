package usecases_test

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestCreatePoolUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		criteria      pairing_value_objects.Criteria
		setupMocks    func(*mocks.MockPoolWriter)
		useWriter     bool
		expectedError string
		validate      func(*testing.T, *pairing_entities.Pool)
	}{
		{
			name: "successfully create pool with pair size 2",
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
			},
			validate: func(t *testing.T, pool *pairing_entities.Pool) {
				assert.NotNil(t, pool)
				assert.Equal(t, uint8(2), pool.PartySize)
				assert.Empty(t, pool.Parties)
			},
		},
		{
			name: "successfully create pool with pair size 5",
			criteria: pairing_value_objects.Criteria{
				PairSize: 5,
			},
			validate: func(t *testing.T, pool *pairing_entities.Pool) {
				assert.NotNil(t, pool)
				assert.Equal(t, uint8(5), pool.PartySize)
			},
		},
		{
			name: "fail with pair size 0",
			criteria: pairing_value_objects.Criteria{
				PairSize: 0,
			},
			expectedError: "pair size must be greater than 0",
		},
		{
			name: "fail with negative pair size",
			criteria: pairing_value_objects.Criteria{
				PairSize: -1,
			},
			expectedError: "pair size must be greater than 0",
		},
		{
			name: "successfully create and persist pool via writer",
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
			},
			useWriter: true,
			setupMocks: func(poolWriter *mocks.MockPoolWriter) {
				poolWriter.On("Save", mock.AnythingOfType("*entities.Pool")).Return(mock.AnythingOfType("*entities.Pool"), nil)
			},
			validate: func(t *testing.T, pool *pairing_entities.Pool) {
				assert.NotNil(t, pool)
				assert.Equal(t, uint8(2), pool.PartySize)
			},
		},
		{
			name: "fail when pool writer returns error",
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
			},
			useWriter: true,
			setupMocks: func(poolWriter *mocks.MockPoolWriter) {
				poolWriter.On("Save", mock.AnythingOfType("*entities.Pool")).Return(nil, assert.AnError)
			},
			expectedError: "failed to save pool",
		},
		{
			name: "pool supports join after creation",
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
			},
			validate: func(t *testing.T, pool *pairing_entities.Pool) {
				assert.NotNil(t, pool)

				party1 := uuid.New()
				party2 := uuid.New()

				pos1 := pool.Join(party1)
				assert.Equal(t, 1, pos1)

				pos2 := pool.Join(party2)
				assert.Equal(t, 2, pos2)

				// Idempotent join
				pos1Again := pool.Join(party1)
				assert.Equal(t, 1, pos1Again)
			},
		},
		{
			name: "pool supports concurrent joins",
			criteria: pairing_value_objects.Criteria{
				PairSize: 10,
			},
			validate: func(t *testing.T, pool *pairing_entities.Pool) {
				assert.NotNil(t, pool)

				var wg sync.WaitGroup
				partyCount := 20
				for i := 0; i < partyCount; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						pid := uuid.New()
						pos := pool.Join(pid)
						assert.Greater(t, pos, 0)
					}()
				}
				wg.Wait()

				assert.Equal(t, partyCount, len(pool.Parties))
			},
		},
		{
			name: "create pool with game criteria",
			criteria: pairing_value_objects.Criteria{
				PairSize: 2,
				GameID:   uuidPtr(uuid.New()),
				Tier:     "gold",
			},
			validate: func(t *testing.T, pool *pairing_entities.Pool) {
				assert.NotNil(t, pool)
				assert.Equal(t, uint8(2), pool.PartySize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPoolWriter := new(mocks.MockPoolWriter)

			if tt.setupMocks != nil {
				tt.setupMocks(mockPoolWriter)
			}

			uc := &usecases.CreatePoolUseCase{}
			if tt.useWriter {
				uc.PoolWriter = mockPoolWriter
			}

			pool, err := uc.Execute(tt.criteria)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, pool)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pool)
				if tt.validate != nil {
					tt.validate(t, pool)
				}
			}

			mockPoolWriter.AssertExpectations(t)
		})
	}
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	return &id
}
