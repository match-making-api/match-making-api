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

func TestUpdateInvitationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		invitationID  uuid.UUID
		payload       usecases.UpdateInvitationPayload
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortInvitationReader, *mocks.MockPortInvitationWriter, *pairing_entities.Invitation)
		expectedError string
		validate      func(*testing.T, *pairing_entities.Invitation)
	}{
		{
			name:         "successfully update invitation message",
			invitationID: uuid.New(),
			payload: usecases.UpdateInvitationPayload{
				Message: func() *string { msg := "Updated message"; return &msg }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				inv.Message = "Original message"
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				writer.On("Save", mock.Anything, mock.MatchedBy(func(savedInv *pairing_entities.Invitation) bool {
					return savedInv.Message == "Updated message"
				})).Return(inv, nil)
			},
			validate: func(t *testing.T, inv *pairing_entities.Invitation) {
				assert.Equal(t, "Updated message", inv.Message)
			},
		},
		{
			name:         "successfully update expiration date",
			invitationID: uuid.New(),
			payload: usecases.UpdateInvitationPayload{
				ExpirationDate: func() *time.Time {
					t := time.Now().Add(48 * time.Hour)
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
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.Invitation")).Return(inv, nil)
			},
		},
		{
			name:         "successfully update both message and expiration date",
			invitationID: uuid.New(),
			payload: usecases.UpdateInvitationPayload{
				Message: func() *string { msg := "Updated message"; return &msg }(),
				ExpirationDate: func() *time.Time {
					t := time.Now().Add(48 * time.Hour)
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
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.Invitation")).Return(inv, nil)
			},
		},
		{
			name:         "fail when user is not admin",
			invitationID: uuid.New(),
			payload: usecases.UpdateInvitationPayload{
				Message: func() *string { msg := "Updated message"; return &msg }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				// No AudienceKey set, so not admin
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, inv *pairing_entities.Invitation) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "only administrators can update invitations",
		},
		{
			name:         "fail when invitation does not exist",
			invitationID: uuid.New(),
			payload: usecases.UpdateInvitationPayload{
				Message: func() *string { msg := "Updated message"; return &msg }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, inv *pairing_entities.Invitation) {
				reader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("invitation not found"))
			},
			expectedError: "failed to get invitation",
		},
		{
			name:         "fail when invitation is not pending",
			invitationID: uuid.New(),
			payload: usecases.UpdateInvitationPayload{
				Message: func() *string { msg := "Updated message"; return &msg }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusAccepted
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			expectedError: "cannot be updated",
		},
		{
			name:         "fail when expiration date is in the past",
			invitationID: uuid.New(),
			payload: usecases.UpdateInvitationPayload{
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
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			expectedError: "expiration date must be in the future",
		},
		{
			name:         "fail when repository returns error on save",
			invitationID: uuid.New(),
			payload: usecases.UpdateInvitationPayload{
				Message: func() *string { msg := "Updated message"; return &msg }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.Invitation")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to update invitation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockPortInvitationReader)
			mockWriter := new(mocks.MockPortInvitationWriter)

			invitation := &pairing_entities.Invitation{
				BaseEntity: common.BaseEntity{ID: tt.invitationID},
			}
			tt.setupMocks(mockReader, mockWriter, invitation)

			useCase := &usecases.UpdateInvitationUseCase{
				InvitationReader: mockReader,
				InvitationWriter: mockWriter,
			}

			ctx := tt.setupContext(context.Background())
			result, err := useCase.Execute(ctx, tt.invitationID, tt.payload)

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

			mockReader.AssertExpectations(t)
			mockWriter.AssertExpectations(t)
		})
	}
}
