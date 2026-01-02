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

func TestResendExternalInvitationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		invitationID  uuid.UUID
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortExternalInvitationReader, *mocks.MockPortExternalInvitationWriter, *mocks.MockExternalInvitationNotifier, *pairing_entities.ExternalInvitation)
		expectedError string
	}{
		{
			name:         "successfully resend external invitation",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, writer *mocks.MockPortExternalInvitationWriter, notifier *mocks.MockExternalInvitationNotifier, inv *pairing_entities.ExternalInvitation) {
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				notifier.On("NotifyInvitationCreated", mock.Anything, inv.ID, inv.Email, inv.FullName, inv.Message, inv.RegistrationToken, inv.MatchID, inv.EventID).Return(nil)
			},
		},
		{
			name:         "fail when user is not admin",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				// No AudienceKey set, so not admin
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, writer *mocks.MockPortExternalInvitationWriter, notifier *mocks.MockExternalInvitationNotifier, inv *pairing_entities.ExternalInvitation) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "only administrators can resend external invitations",
		},
		{
			name:         "fail when invitation not found",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, writer *mocks.MockPortExternalInvitationWriter, notifier *mocks.MockExternalInvitationNotifier, inv *pairing_entities.ExternalInvitation) {
				reader.On("GetByID", mock.Anything, inv.ID).Return(nil, errors.New("not found"))
			},
			expectedError: "failed to get external invitation",
		},
		{
			name:         "fail when invitation is not pending",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, writer *mocks.MockPortExternalInvitationWriter, notifier *mocks.MockExternalInvitationNotifier, inv *pairing_entities.ExternalInvitation) {
				inv.Status = pairing_entities.ExternalInvitationStatusAccepted
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			expectedError: "can only resend pending invitations",
		},
		{
			name:         "fail when invitation is expired",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, writer *mocks.MockPortExternalInvitationWriter, notifier *mocks.MockExternalInvitationNotifier, inv *pairing_entities.ExternalInvitation) {
				expiredTime := time.Now().Add(-1 * time.Hour)
				inv.ExpirationDate = &expiredTime
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			expectedError: "cannot resend expired invitation",
		},
		{
			name:         "fail when notification fails",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, writer *mocks.MockPortExternalInvitationWriter, notifier *mocks.MockExternalInvitationNotifier, inv *pairing_entities.ExternalInvitation) {
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				notifier.On("NotifyInvitationCreated", mock.Anything, inv.ID, inv.Email, inv.FullName, inv.Message, inv.RegistrationToken, inv.MatchID, inv.EventID).Return(errors.New("email service error"))
			},
			expectedError: "failed to resend invitation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockPortExternalInvitationReader)
			mockWriter := new(mocks.MockPortExternalInvitationWriter)
			mockNotifier := new(mocks.MockExternalInvitationNotifier)
			inv := &pairing_entities.ExternalInvitation{
				BaseEntity:        common.BaseEntity{ID: tt.invitationID},
				FullName:          "John Doe",
				Email:             "john.doe@example.com",
				Message:          "Test message",
				RegistrationToken: "test-token",
				Status:            pairing_entities.ExternalInvitationStatusPending,
				ExpirationDate: func() *time.Time {
					t := time.Now().Add(24 * time.Hour)
					return &t
				}(),
			}
			tt.setupMocks(mockReader, mockWriter, mockNotifier, inv)

			useCase := &usecases.ResendExternalInvitationUseCase{
				ExternalInvitationReader: mockReader,
				ExternalInvitationWriter:  mockWriter,
				Notifier:                  mockNotifier,
			}

			ctx := tt.setupContext(context.Background())
			err := useCase.Execute(ctx, tt.invitationID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockReader.AssertExpectations(t)
			mockWriter.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}
