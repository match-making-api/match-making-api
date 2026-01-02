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

func TestRevokeInvitationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		invitationID  uuid.UUID
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortInvitationReader, *mocks.MockPortInvitationWriter, *mocks.MockInvitationNotifier, *pairing_entities.Invitation)
		expectedError string
	}{
		{
			name:         "successfully revoke invitation",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				writer.On("Save", mock.Anything, mock.MatchedBy(func(savedInv *pairing_entities.Invitation) bool {
					return savedInv.Status == pairing_entities.InvitationStatusRevoked && savedInv.RevokedAt != nil && savedInv.RevokedBy != nil
				})).Return(inv, nil)
				notifier.On("NotifyInvitationRevoked", mock.Anything, inv.ID, inv.UserID).Return(nil)
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
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "only administrators can revoke invitations",
		},
		{
			name:         "fail when invitation does not exist",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				reader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("invitation not found"))
			},
			expectedError: "failed to get invitation",
		},
		{
			name:         "fail when invitation cannot be revoked (already accepted)",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusAccepted
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			expectedError: "cannot be revoked",
		},
		{
			name:         "fail when repository returns error on save",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.Invitation")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to revoke invitation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockPortInvitationReader)
			mockWriter := new(mocks.MockPortInvitationWriter)
			mockNotifier := new(mocks.MockInvitationNotifier)

			invitation := &pairing_entities.Invitation{
				BaseEntity: common.BaseEntity{ID: tt.invitationID},
				UserID:     uuid.New(),
			}
			tt.setupMocks(mockReader, mockWriter, mockNotifier, invitation)

			useCase := &usecases.RevokeInvitationUseCase{
				InvitationReader: mockReader,
				InvitationWriter: mockWriter,
				Notifier:         mockNotifier,
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
