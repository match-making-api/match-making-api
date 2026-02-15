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
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestCreateExternalInvitationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		payload       usecases.CreateExternalInvitationPayload
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortExternalInvitationWriter, *mocks.MockPortExternalInvitationReader, *mocks.MockPortPairReader, *mocks.MockExternalInvitationNotifier)
		expectedError string
		validate      func(*testing.T, *pairing_entities.ExternalInvitation)
	}{
		{
			name: "successfully create match external invitation",
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "John Doe",
				Email:    "john.doe@example.com",
				Message:  "You are invited to join a match",
				ExpirationDate: func() *time.Time {
					t := time.Now().Add(24 * time.Hour)
					return &t
				}(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// Check for existing invitations
				reader.On("FindByEmail", mock.Anything, "john.doe@example.com").Return([]*pairing_entities.ExternalInvitation{}, nil)
				// Validate match
				matchID := uuid.New()
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: matchID},
					ConflictStatus: pairing_entities.ConflictStatusNone,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
				// Save invitation
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.ExternalInvitation")).Run(func(args mock.Arguments) {
					inv := args.Get(1).(*pairing_entities.ExternalInvitation)
					if inv.ID == uuid.Nil {
						inv.ID = uuid.New()
					}
				}).Return(mock.Anything, nil)
				// Send notification
				notifier.On("NotifyInvitationCreated", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
			},
			validate: func(t *testing.T, inv *pairing_entities.ExternalInvitation) {
				assert.NotEqual(t, uuid.Nil, inv.ID)
				assert.Equal(t, pairing_entities.ExternalInvitationStatusPending, inv.Status)
				assert.Equal(t, "John Doe", inv.FullName)
				assert.Equal(t, "john.doe@example.com", inv.Email)
				assert.NotEmpty(t, inv.RegistrationToken)
				assert.NotNil(t, inv.MatchID)
			},
		},
		{
			name: "fail when user is not admin",
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "John Doe",
				Email:    "john.doe@example.com",
				Message:  "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				// No AudienceKey set, so not admin
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "only administrators can create external invitations",
		},
		{
			name: "fail when email is invalid",
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "John Doe",
				Email:    "invalid-email",
				Message:  "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "email validation failed",
		},
		{
			name: "fail when full name is empty",
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "",
				Email:    "john.doe@example.com",
				Message:  "Test message",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "full name is required",
		},
		{
			name: "fail when match_id is missing for match invitation",
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "John Doe",
				Email:    "john.doe@example.com",
				Message:  "Test message",
				MatchID:   nil,
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// No mocks needed: validateMatchOrEvent fails before FindByEmail is called
			},
			expectedError: "match_id is required for match invitations",
		},
		{
			name: "fail when match does not exist",
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "John Doe",
				Email:    "john.doe@example.com",
				Message:  "Test message",
				MatchID:  func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// validateMatchOrEvent is called before FindByEmail
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("match not found"))
			},
			expectedError: "match/event validation failed",
		},
		{
			name: "fail when match has conflicts",
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "John Doe",
				Email:    "john.doe@example.com",
				Message:  "Test message",
				MatchID:  func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// validateMatchOrEvent is called before FindByEmail
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
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "John Doe",
				Email:    "john.doe@example.com",
				Message:  "Test message",
				ExpirationDate: func() *time.Time {
					t := time.Now().Add(-1 * time.Hour)
					return &t
				}(),
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// validateMatchOrEvent is called before expiration check, so mock it to pass
				matchID := uuid.New()
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: matchID},
					ConflictStatus: pairing_entities.ConflictStatusNone,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
			},
			expectedError: "expiration date must be in the future",
		},
		{
			name: "fail when pending invitation already exists",
			payload: func() usecases.CreateExternalInvitationPayload {
				matchID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
				return usecases.CreateExternalInvitationPayload{
					Type:     pairing_entities.ExternalInvitationTypeMatch,
					FullName: "John Doe",
					Email:    "john.doe@example.com",
					Message:  "Test message",
					MatchID:  &matchID,
				}
			}(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				// validateMatchOrEvent is called before FindByEmail, so mock PairReader
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: uuid.New()},
					ConflictStatus: pairing_entities.ConflictStatusNone,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
				// Return existing pending invitation with the same MatchID
				matchID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
				existingInv := &pairing_entities.ExternalInvitation{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Email:      "john.doe@example.com",
					MatchID:    &matchID,
					Status:     pairing_entities.ExternalInvitationStatusPending,
					ExpirationDate: func() *time.Time {
						t := time.Now().Add(24 * time.Hour)
						return &t
					}(),
				}
				reader.On("FindByEmail", mock.Anything, "john.doe@example.com").Return([]*pairing_entities.ExternalInvitation{existingInv}, nil)
			},
			expectedError: "a pending invitation already exists",
		},
		{
			name: "fail when repository returns error",
			payload: usecases.CreateExternalInvitationPayload{
				Type:     pairing_entities.ExternalInvitationTypeMatch,
				FullName: "John Doe",
				Email:    "john.doe@example.com",
				Message:  "Test message",
				MatchID:  func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortExternalInvitationWriter, reader *mocks.MockPortExternalInvitationReader, pairReader *mocks.MockPortPairReader, notifier *mocks.MockExternalInvitationNotifier) {
				reader.On("FindByEmail", mock.Anything, "john.doe@example.com").Return([]*pairing_entities.ExternalInvitation{}, nil)
				matchID := uuid.New()
				pair := &pairing_entities.Pair{
					BaseEntity:     common.BaseEntity{ID: matchID},
					ConflictStatus: pairing_entities.ConflictStatusNone,
				}
				pairReader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(pair, nil)
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.ExternalInvitation")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to create invitation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(mocks.MockPortExternalInvitationWriter)
			mockReader := new(mocks.MockPortExternalInvitationReader)
			mockPairReader := new(mocks.MockPortPairReader)
			mockNotifier := new(mocks.MockExternalInvitationNotifier)
			tt.setupMocks(mockWriter, mockReader, mockPairReader, mockNotifier)

			useCase := &usecases.CreateExternalInvitationUseCase{
				ExternalInvitationWriter: mockWriter,
				ExternalInvitationReader:  mockReader,
				PairReader:               mockPairReader,
				Notifier:                 mockNotifier,
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
			mockReader.AssertExpectations(t)
			mockPairReader.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}
