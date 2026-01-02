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

func TestListExternalInvitationsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		filter        usecases.ListExternalInvitationsFilter
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortExternalInvitationReader)
		expectedError string
		validate      func(*testing.T, []*pairing_entities.ExternalInvitation)
	}{
		{
			name: "successfully list by email",
			filter: usecases.ListExternalInvitationsFilter{
				Email: func() *string { s := "john.doe@example.com"; return &s }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader) {
				invitations := []*pairing_entities.ExternalInvitation{
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Email:      "john.doe@example.com",
						Status:     pairing_entities.ExternalInvitationStatusPending,
					},
				}
				reader.On("FindByEmail", mock.Anything, "john.doe@example.com").Return(invitations, nil)
			},
			validate: func(t *testing.T, invs []*pairing_entities.ExternalInvitation) {
				assert.Len(t, invs, 1)
				assert.Equal(t, "john.doe@example.com", invs[0].Email)
			},
		},
		{
			name: "successfully list by match_id",
			filter: usecases.ListExternalInvitationsFilter{
				MatchID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader) {
				matchID := uuid.New()
				invitations := []*pairing_entities.ExternalInvitation{
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						MatchID:    &matchID,
						Status:     pairing_entities.ExternalInvitationStatusPending,
					},
				}
				reader.On("FindByMatchID", mock.Anything, matchID).Return(invitations, nil)
			},
			validate: func(t *testing.T, invs []*pairing_entities.ExternalInvitation) {
				assert.Len(t, invs, 1)
				assert.NotNil(t, invs[0].MatchID)
			},
		},
		{
			name: "successfully filter by status",
			filter: usecases.ListExternalInvitationsFilter{
				Email: func() *string { s := "john.doe@example.com"; return &s }(),
				Status: func() *pairing_entities.ExternalInvitationStatus {
					s := pairing_entities.ExternalInvitationStatusPending
					return &s
				}(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader) {
				invitations := []*pairing_entities.ExternalInvitation{
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Email:      "john.doe@example.com",
						Status:     pairing_entities.ExternalInvitationStatusPending,
					},
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Email:      "john.doe@example.com",
						Status:     pairing_entities.ExternalInvitationStatusAccepted,
					},
				}
				reader.On("FindByEmail", mock.Anything, "john.doe@example.com").Return(invitations, nil)
			},
			validate: func(t *testing.T, invs []*pairing_entities.ExternalInvitation) {
				assert.Len(t, invs, 1)
				assert.Equal(t, pairing_entities.ExternalInvitationStatusPending, invs[0].Status)
			},
		},
		{
			name: "fail when user is not admin",
			filter: usecases.ListExternalInvitationsFilter{
				Email: func() *string { s := "john.doe@example.com"; return &s }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				// No AudienceKey set, so not admin
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "only administrators can list external invitations",
		},
		{
			name:   "fail when no filter provided",
			filter: usecases.ListExternalInvitationsFilter{},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "at least one filter parameter is required",
		},
		{
			name: "fail when repository returns error",
			filter: usecases.ListExternalInvitationsFilter{
				Email: func() *string { s := "john.doe@example.com"; return &s }(),
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortExternalInvitationReader) {
				reader.On("FindByEmail", mock.Anything, "john.doe@example.com").Return(nil, errors.New("database error"))
			},
			expectedError: "failed to list external invitations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockPortExternalInvitationReader)
			tt.setupMocks(mockReader)

			useCase := &usecases.ListExternalInvitationsUseCase{
				ExternalInvitationReader: mockReader,
			}

			ctx := tt.setupContext(context.Background())
			result, err := useCase.Execute(ctx, tt.filter)

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
