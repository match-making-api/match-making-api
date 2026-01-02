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

func TestGetExternalInvitationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		invitationID  uuid.UUID
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortExternalInvitationReader, *pairing_entities.ExternalInvitation)
		expectedError string
		validate      func(*testing.T, *pairing_entities.ExternalInvitation)
	}{
		{
			name:         "successfully get external invitation",
			invitationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, inv *pairing_entities.ExternalInvitation) {
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			validate: func(t *testing.T, inv *pairing_entities.ExternalInvitation) {
				assert.NotEqual(t, uuid.Nil, inv.ID)
				assert.Equal(t, "John Doe", inv.FullName)
				assert.Equal(t, "john.doe@example.com", inv.Email)
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
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, inv *pairing_entities.ExternalInvitation) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "only administrators can view external invitations",
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
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, inv *pairing_entities.ExternalInvitation) {
				reader.On("GetByID", mock.Anything, inv.ID).Return(nil, errors.New("not found"))
			},
			expectedError: "failed to get external invitation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockPortExternalInvitationReader)
			inv := &pairing_entities.ExternalInvitation{
				BaseEntity: common.BaseEntity{ID: tt.invitationID},
				FullName:   "John Doe",
				Email:      "john.doe@example.com",
				Status:     pairing_entities.ExternalInvitationStatusPending,
			}
			tt.setupMocks(mockReader, inv)

			useCase := &usecases.GetExternalInvitationUseCase{
				ExternalInvitationReader: mockReader,
			}

			ctx := tt.setupContext(context.Background())
			result, err := useCase.Execute(ctx, tt.invitationID)

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
		})
	}
}

func TestGetExternalInvitationByTokenUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		token           string
		setupMocks     func(*mocks.MockPortExternalInvitationReader, *pairing_entities.ExternalInvitation)
		expectedError  string
		validate       func(*testing.T, *pairing_entities.ExternalInvitation)
	}{
		{
			name:  "successfully get external invitation by token",
			token:  "test-token-123",
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, inv *pairing_entities.ExternalInvitation) {
				reader.On("GetByRegistrationToken", mock.Anything, "test-token-123").Return(inv, nil)
			},
			validate: func(t *testing.T, inv *pairing_entities.ExternalInvitation) {
				assert.NotEqual(t, uuid.Nil, inv.ID)
				assert.Equal(t, "test-token-123", inv.RegistrationToken)
			},
		},
		{
			name:  "fail when token is empty",
			token:  "",
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, inv *pairing_entities.ExternalInvitation) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "registration token is required",
		},
		{
			name:  "fail when invitation not found",
			token:  "invalid-token",
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader, inv *pairing_entities.ExternalInvitation) {
				reader.On("GetByRegistrationToken", mock.Anything, "invalid-token").Return(nil, errors.New("not found"))
			},
			expectedError: "failed to get external invitation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockPortExternalInvitationReader)
			inv := &pairing_entities.ExternalInvitation{
				BaseEntity:        common.BaseEntity{ID: uuid.New()},
				RegistrationToken: tt.token,
				FullName:          "John Doe",
				Email:             "john.doe@example.com",
				Status:            pairing_entities.ExternalInvitationStatusPending,
			}
			tt.setupMocks(mockReader, inv)

			useCase := &usecases.GetExternalInvitationByTokenUseCase{
				ExternalInvitationReader: mockReader,
			}

			result, err := useCase.Execute(context.Background(), tt.token)

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
		})
	}
}
