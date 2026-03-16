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
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func adminContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
	return ctx
}

func nonAdminContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
	return ctx
}

func TestResolveMatchConflictUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		payload       usecases.ResolveConflictPayload
		useAdminCtx   bool
		setupMocks    func(*mocks.MockPortPairReader, *mocks.MockPortPairWriter)
		expectedError string
	}{
		{
			name: "successfully resolve a flagged conflict",
			payload: usecases.ResolveConflictPayload{
				PairID: uuid.New(),
				Action: usecases.ConflictResolutionResolve,
				Reason: "Schedule adjusted",
			},
			useAdminCtx: true,
			setupMocks: func(pairReader *mocks.MockPortPairReader, pairWriter *mocks.MockPortPairWriter) {
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: uuid.New()},
					ConflictStatus: pairing_entities.ConflictStatusFlagged,
					ConflictReason: "overlapping schedules",
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
				pairWriter.On("Save", mock.MatchedBy(func(p *pairing_entities.Pair) bool {
					return p.ConflictStatus == pairing_entities.ConflictStatusResolved && p.ConflictReason == ""
				})).Return(pair, nil)
			},
		},
		{
			name: "successfully override a flagged conflict",
			payload: usecases.ResolveConflictPayload{
				PairID: uuid.New(),
				Action: usecases.ConflictResolutionOverride,
				Reason: "Admin override",
			},
			useAdminCtx: true,
			setupMocks: func(pairReader *mocks.MockPortPairReader, pairWriter *mocks.MockPortPairWriter) {
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: uuid.New()},
					ConflictStatus: pairing_entities.ConflictStatusFlagged,
					ConflictReason: "schedule conflict",
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
				pairWriter.On("Save", mock.MatchedBy(func(p *pairing_entities.Pair) bool {
					return p.ConflictStatus == pairing_entities.ConflictStatusNone && p.ConflictReason == ""
				})).Return(pair, nil)
			},
		},
		{
			name: "fail when not admin",
			payload: usecases.ResolveConflictPayload{
				PairID: uuid.New(),
				Action: usecases.ConflictResolutionResolve,
			},
			useAdminCtx: false,
			setupMocks: func(_ *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter) {
				// No mocks needed — fails at admin check
			},
			expectedError: "only administrators can resolve conflicts",
		},
		{
			name: "fail when pair not found",
			payload: usecases.ResolveConflictPayload{
				PairID: uuid.New(),
				Action: usecases.ConflictResolutionResolve,
			},
			useAdminCtx: true,
			setupMocks: func(pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter) {
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("pair not found"))
			},
			expectedError: "failed to get pair",
		},
		{
			name: "fail when pair is not flagged",
			payload: usecases.ResolveConflictPayload{
				PairID: uuid.New(),
				Action: usecases.ConflictResolutionResolve,
			},
			useAdminCtx: true,
			setupMocks: func(pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter) {
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: uuid.New()},
					ConflictStatus: pairing_entities.ConflictStatusNone,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
			},
			expectedError: "not flagged as conflicting",
		},
		{
			name: "fail with remove action (not implemented)",
			payload: usecases.ResolveConflictPayload{
				PairID: uuid.New(),
				Action: usecases.ConflictResolutionRemove,
			},
			useAdminCtx: true,
			setupMocks: func(pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter) {
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: uuid.New()},
					ConflictStatus: pairing_entities.ConflictStatusFlagged,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
			},
			expectedError: "not yet implemented",
		},
		{
			name: "fail with invalid action",
			payload: usecases.ResolveConflictPayload{
				PairID: uuid.New(),
				Action: "invalid_action",
			},
			useAdminCtx: true,
			setupMocks: func(pairReader *mocks.MockPortPairReader, _ *mocks.MockPortPairWriter) {
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: uuid.New()},
					ConflictStatus: pairing_entities.ConflictStatusFlagged,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
			},
			expectedError: "invalid conflict resolution action",
		},
		{
			name: "fail when save returns error",
			payload: usecases.ResolveConflictPayload{
				PairID: uuid.New(),
				Action: usecases.ConflictResolutionResolve,
			},
			useAdminCtx: true,
			setupMocks: func(pairReader *mocks.MockPortPairReader, pairWriter *mocks.MockPortPairWriter) {
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: uuid.New()},
					ConflictStatus: pairing_entities.ConflictStatusFlagged,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
				pairWriter.On("Save", mock.AnythingOfType("*entities.Pair")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to save resolved pair",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPairReader := new(mocks.MockPortPairReader)
			mockPairWriter := new(mocks.MockPortPairWriter)

			tt.setupMocks(mockPairReader, mockPairWriter)

			useCase := &usecases.ResolveMatchConflictUseCase{
				PairReader: mockPairReader,
				PairWriter: mockPairWriter,
			}

			var ctx context.Context
			if tt.useAdminCtx {
				ctx = adminContext()
			} else {
				ctx = nonAdminContext()
			}

			err := useCase.Execute(ctx, tt.payload)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockPairReader.AssertExpectations(t)
			mockPairWriter.AssertExpectations(t)
		})
	}
}
