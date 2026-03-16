package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	parties_entities "github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func newContextWithResourceOwner() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
	return ctx
}

func TestCreatePairUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		partyIDs      []uuid.UUID
		setupMocks    func(*mocks.MockPortPartyReader, *mocks.MockPortPairWriter, *mocks.MockConflictVerifier)
		withConflict  bool
		expectedError string
		validate      func(*testing.T, *pairing_entities.Pair)
	}{
		{
			name:     "successfully create pair with two parties",
			partyIDs: []uuid.UUID{uuid.New(), uuid.New()},
			setupMocks: func(partyReader *mocks.MockPortPartyReader, pairWriter *mocks.MockPortPairWriter, _ *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Party{ID: uuid.New()}, nil).Times(2)
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Run(func(args mock.Arguments) {
					pair := args.Get(0).(*pairing_entities.Pair)
					pair.ID = uuid.New()
				}).Return(mock.AnythingOfType("*entities.Pair"), nil)
			},
			validate: func(t *testing.T, pair *pairing_entities.Pair) {
				assert.NotNil(t, pair)
				assert.Equal(t, 2, len(pair.Match))
				assert.Equal(t, pairing_entities.ConflictStatusNone, pair.ConflictStatus)
			},
		},
		{
			name:     "successfully create pair with three parties",
			partyIDs: []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
			setupMocks: func(partyReader *mocks.MockPortPartyReader, pairWriter *mocks.MockPortPairWriter, _ *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Party{ID: uuid.New()}, nil).Times(3)
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Run(func(args mock.Arguments) {
					pair := args.Get(0).(*pairing_entities.Pair)
					pair.ID = uuid.New()
				}).Return(mock.AnythingOfType("*entities.Pair"), nil)
			},
			validate: func(t *testing.T, pair *pairing_entities.Pair) {
				assert.NotNil(t, pair)
				assert.Equal(t, 3, len(pair.Match))
			},
		},
		{
			name:     "fail when party not found",
			partyIDs: []uuid.UUID{uuid.New(), uuid.New()},
			setupMocks: func(partyReader *mocks.MockPortPartyReader, _ *mocks.MockPortPairWriter, _ *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("party not found")).Once()
			},
			expectedError: "not found",
		},
		{
			name:     "fail when first party ok but second party not found",
			partyIDs: []uuid.UUID{uuid.New(), uuid.New()},
			setupMocks: func(partyReader *mocks.MockPortPartyReader, _ *mocks.MockPortPairWriter, _ *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Party{ID: uuid.New()}, nil).Once()
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("party not found")).Once()
			},
			expectedError: "not found",
		},
		{
			name:     "fail when pair writer returns error",
			partyIDs: []uuid.UUID{uuid.New(), uuid.New()},
			setupMocks: func(partyReader *mocks.MockPortPartyReader, pairWriter *mocks.MockPortPairWriter, _ *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Party{ID: uuid.New()}, nil).Times(2)
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Return(nil, errors.New("database error"))
			},
			expectedError: "create error",
		},
		{
			name:         "successfully create pair with conflict verification - no conflicts",
			partyIDs:     []uuid.UUID{uuid.New(), uuid.New()},
			withConflict: true,
			setupMocks: func(partyReader *mocks.MockPortPartyReader, pairWriter *mocks.MockPortPairWriter, conflictVerifier *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Party{ID: uuid.New()}, nil).Times(2)
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Run(func(args mock.Arguments) {
					pair := args.Get(0).(*pairing_entities.Pair)
					pair.ID = uuid.New()
				}).Return(mock.AnythingOfType("*entities.Pair"), nil)
				conflictVerifier.On("Execute", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(
					&usecases.ConflictResult{HasConflict: false, ConflictingPairs: []uuid.UUID{}}, nil,
				)
			},
			validate: func(t *testing.T, pair *pairing_entities.Pair) {
				assert.NotNil(t, pair)
				assert.Equal(t, 2, len(pair.Match))
			},
		},
		{
			name:         "successfully create pair even when conflict verification finds conflicts",
			partyIDs:     []uuid.UUID{uuid.New(), uuid.New()},
			withConflict: true,
			setupMocks: func(partyReader *mocks.MockPortPartyReader, pairWriter *mocks.MockPortPairWriter, conflictVerifier *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Party{ID: uuid.New()}, nil).Times(2)
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Run(func(args mock.Arguments) {
					pair := args.Get(0).(*pairing_entities.Pair)
					pair.ID = uuid.New()
				}).Return(mock.AnythingOfType("*entities.Pair"), nil)
				conflictVerifier.On("Execute", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(
					&usecases.ConflictResult{HasConflict: true, ConflictingPairs: []uuid.UUID{uuid.New()}}, nil,
				)
			},
			validate: func(t *testing.T, pair *pairing_entities.Pair) {
				assert.NotNil(t, pair)
				assert.Equal(t, 2, len(pair.Match))
			},
		},
		{
			name:         "successfully create pair when conflict verifier returns error",
			partyIDs:     []uuid.UUID{uuid.New(), uuid.New()},
			withConflict: true,
			setupMocks: func(partyReader *mocks.MockPortPartyReader, pairWriter *mocks.MockPortPairWriter, conflictVerifier *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Party{ID: uuid.New()}, nil).Times(2)
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Run(func(args mock.Arguments) {
					pair := args.Get(0).(*pairing_entities.Pair)
					pair.ID = uuid.New()
				}).Return(mock.AnythingOfType("*entities.Pair"), nil)
				conflictVerifier.On("Execute", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(
					nil, errors.New("conflict service unavailable"),
				)
			},
			validate: func(t *testing.T, pair *pairing_entities.Pair) {
				assert.NotNil(t, pair)
				assert.Equal(t, 2, len(pair.Match))
			},
		},
		{
			name:     "successfully create pair without conflict verifier (nil)",
			partyIDs: []uuid.UUID{uuid.New(), uuid.New()},
			setupMocks: func(partyReader *mocks.MockPortPartyReader, pairWriter *mocks.MockPortPairWriter, _ *mocks.MockConflictVerifier) {
				partyReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Party{ID: uuid.New()}, nil).Times(2)
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Run(func(args mock.Arguments) {
					pair := args.Get(0).(*pairing_entities.Pair)
					pair.ID = uuid.New()
				}).Return(mock.AnythingOfType("*entities.Pair"), nil)
			},
			validate: func(t *testing.T, pair *pairing_entities.Pair) {
				assert.NotNil(t, pair)
				assert.Equal(t, 2, len(pair.Match))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartyReader := new(mocks.MockPortPartyReader)
			mockPairWriter := new(mocks.MockPortPairWriter)
			mockConflictVerifier := new(mocks.MockConflictVerifier)

			tt.setupMocks(mockPartyReader, mockPairWriter, mockConflictVerifier)

			useCase := &usecases.CreatePairUseCase{
				PartyReader: mockPartyReader,
				PairWriter:  mockPairWriter,
			}

			if tt.withConflict {
				useCase.ConflictVerifier = mockConflictVerifier
			}

			ctx := newContextWithResourceOwner()
			result, err := useCase.Execute(ctx, tt.partyIDs)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			mockPartyReader.AssertExpectations(t)
			mockPairWriter.AssertExpectations(t)
			if tt.withConflict {
				mockConflictVerifier.AssertExpectations(t)
			}
		})
	}
}
