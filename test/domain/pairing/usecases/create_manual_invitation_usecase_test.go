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
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestCreateManualInvitationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		payload       usecases.CreateInvitationPayload
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortInvitationWriter, *mocks.MockPortPeerReader, *mocks.MockPortPairReader, *mocks.MockInvitationNotifier)
		expectedError string
		validate      func(*testing.T, *pairing_entities.Invitation)
	}{
		{
			name: "successfully create match invitation",
			payload: usecases.CreateInvitationPayload{
				Type:    pairing_entities.InvitationTypeMatch,
				UserID:  uuid.New(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Message: "You are invited to join a match",
				ExpirationDate: func() *time.Time {
					t := time.Now().Add(24 * time.Hour)
					return &t
				}(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortInvitationWriter, peerReader *mocks.MockPortPeerReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockInvitationNotifier) {
				peerReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Peer{ID: uuid.New()}, nil)
				matchID := uuid.New()
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: matchID},
					ConflictStatus: pairing_entities.ConflictStatusNone,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.Invitation")).Return(func(ctx context.Context, inv *pairing_entities.Invitation) *pairing_entities.Invitation {
					inv.ID = uuid.New()
					return inv
				}, nil)
				notifier.On("NotifyInvitationCreated", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("string")).Return(nil)
			},
			validate: func(t *testing.T, inv *pairing_entities.Invitation) {
				assert.NotEqual(t, uuid.Nil, inv.ID)
				assert.Equal(t, pairing_entities.InvitationStatusPending, inv.Status)
				assert.NotNil(t, inv.MatchID)
			},
		},
		{
			name: "fail when user is not admin",
			payload: usecases.CreateInvitationPayload{
				Type:    pairing_entities.InvitationTypeMatch,
				UserID:  uuid.New(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Message: "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				// No AudienceKey set, so not admin
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortInvitationWriter, peerReader *mocks.MockPortPeerReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockInvitationNotifier) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "only administrators can create manual invitations",
		},
		{
			name: "fail when user does not exist",
			payload: usecases.CreateInvitationPayload{
				Type:    pairing_entities.InvitationTypeMatch,
				UserID:  uuid.New(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Message: "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortInvitationWriter, peerReader *mocks.MockPortPeerReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockInvitationNotifier) {
				peerReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("user not found"))
			},
			expectedError: "user validation failed",
		},
		{
			name: "fail when match_id is missing for match invitation",
			payload: usecases.CreateInvitationPayload{
				Type:    pairing_entities.InvitationTypeMatch,
				UserID:  uuid.New(),
				MatchID: nil,
				Message: "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortInvitationWriter, peerReader *mocks.MockPortPeerReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockInvitationNotifier) {
				peerReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Peer{ID: uuid.New()}, nil)
			},
			expectedError: "match_id is required for match invitations",
		},
		{
			name: "fail when match does not exist",
			payload: usecases.CreateInvitationPayload{
				Type:    pairing_entities.InvitationTypeMatch,
				UserID:  uuid.New(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Message: "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortInvitationWriter, peerReader *mocks.MockPortPeerReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockInvitationNotifier) {
				peerReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Peer{ID: uuid.New()}, nil)
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("match not found"))
			},
			expectedError: "match/event validation failed",
		},
		{
			name: "fail when match has conflicts",
			payload: usecases.CreateInvitationPayload{
				Type:    pairing_entities.InvitationTypeMatch,
				UserID:  uuid.New(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Message: "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortInvitationWriter, peerReader *mocks.MockPortPeerReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockInvitationNotifier) {
				peerReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Peer{ID: uuid.New()}, nil)
				matchID := uuid.New()
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: matchID},
					ConflictStatus: pairing_entities.ConflictStatusFlagged,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
			},
			expectedError: "has conflicts and is not open for new participants",
		},
		{
			name: "fail when expiration date is in the past",
			payload: usecases.CreateInvitationPayload{
				Type:    pairing_entities.InvitationTypeMatch,
				UserID:  uuid.New(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Message: "Test message",
				ExpirationDate: func() *time.Time {
					t := time.Now().Add(-1 * time.Hour)
					return &t
				}(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortInvitationWriter, peerReader *mocks.MockPortPeerReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockInvitationNotifier) {
				peerReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Peer{ID: uuid.New()}, nil)
			},
			expectedError: "expiration date must be in the future",
		},
		{
			name: "fail when repository returns error",
			payload: usecases.CreateInvitationPayload{
				Type:    pairing_entities.InvitationTypeMatch,
				UserID:  uuid.New(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Message: "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortInvitationWriter, peerReader *mocks.MockPortPeerReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockInvitationNotifier) {
				peerReader.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(&parties_entities.Peer{ID: uuid.New()}, nil)
				matchID := uuid.New()
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: matchID},
					ConflictStatus: pairing_entities.ConflictStatusNone,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.Invitation")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to create invitation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(mocks.MockPortInvitationWriter)
			mockPeerReader := new(mocks.MockPortPeerReader)
			mockPairReader := new(mocks.MockPortPairReader)
			mockNotifier := new(mocks.MockInvitationNotifier)
			tt.setupMocks(mockWriter, mockPeerReader, mockPairReader, mockNotifier)

			useCase := &usecases.CreateManualInvitationUseCase{
				InvitationWriter: mockWriter,
				PeerReader:       mockPeerReader,
				PairReader:       mockPairReader,
				Notifier:         mockNotifier,
			}

			ctx := tt.setupContext(context.Background())
			result, err := useCase.Execute(ctx, tt.payload)

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

			mockWriter.AssertExpectations(t)
			mockPeerReader.AssertExpectations(t)
			mockPairReader.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}
