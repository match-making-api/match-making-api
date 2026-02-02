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

func TestGetUserNotificationsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		userID        uuid.UUID
		limit         int
		offset        int
		setupContext  func(context.Context, uuid.UUID) context.Context
		setupMocks    func(*mocks.MockPortNotificationReader, uuid.UUID)
		expectedError string
		validate      func(*testing.T, *usecases.GetUserNotificationsResult)
	}{
		{
			name:   "successfully get user notifications",
			userID:  uuid.New(),
			limit:  20,
			offset: 0,
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, userID uuid.UUID) {
				notifications := []*pairing_entities.Notification{
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						UserID:     userID,
						Channel:    pairing_entities.NotificationChannelInApp,
						Type:       pairing_entities.NotificationTypeMatchInvitation,
						Title:      "Match Invitation",
						Message:    "You have been invited",
					},
				}
				reader.On("FindByUserID", mock.Anything, userID, 20, 0).Return(notifications, nil)
				reader.On("CountByUserID", mock.Anything, userID).Return(1, nil)
			},
			validate: func(t *testing.T, result *usecases.GetUserNotificationsResult) {
				assert.NotNil(t, result)
				assert.Len(t, result.Notifications, 1)
				assert.Equal(t, 1, result.Total)
				assert.Equal(t, 20, result.Limit)
				assert.Equal(t, 0, result.Offset)
			},
		},
		{
			name:   "fail when user tries to access other user's notifications",
			userID:  uuid.New(),
			limit:  20,
			offset: 0,
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New()) // Different user
				ctx = context.WithValue(ctx, common.AudienceKey, common.UserAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, userID uuid.UUID) {
				// No mocks needed, should fail before calling them
			},
			expectedError: "user can only access their own notifications",
		},
		{
			name:   "admin can access any user's notifications",
			userID:  uuid.New(),
			limit:  20,
			offset: 0,
			setupContext: func(ctx context.Context, _ uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, userID uuid.UUID) {
				notifications := []*pairing_entities.Notification{
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						UserID:     userID,
						Channel:    pairing_entities.NotificationChannelInApp,
						Type:       pairing_entities.NotificationTypeMatchInvitation,
						Title:      "Match Invitation",
						Message:    "You have been invited",
					},
				}
				reader.On("FindByUserID", mock.Anything, userID, 20, 0).Return(notifications, nil)
				reader.On("CountByUserID", mock.Anything, userID).Return(1, nil)
			},
			validate: func(t *testing.T, result *usecases.GetUserNotificationsResult) {
				assert.NotNil(t, result)
				assert.Len(t, result.Notifications, 1)
			},
		},
		{
			name:   "use default limit when limit is 0",
			userID:  uuid.New(),
			limit:  0,
			offset: 0,
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, userID uuid.UUID) {
				reader.On("FindByUserID", mock.Anything, userID, 20, 0).Return([]*pairing_entities.Notification{}, nil)
				reader.On("CountByUserID", mock.Anything, userID).Return(0, nil)
			},
			validate: func(t *testing.T, result *usecases.GetUserNotificationsResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 20, result.Limit)
			},
		},
		{
			name:   "cap limit at 100",
			userID:  uuid.New(),
			limit:  150,
			offset: 0,
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, userID uuid.UUID) {
				reader.On("FindByUserID", mock.Anything, userID, 100, 0).Return([]*pairing_entities.Notification{}, nil)
				reader.On("CountByUserID", mock.Anything, userID).Return(0, nil)
			},
			validate: func(t *testing.T, result *usecases.GetUserNotificationsResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 100, result.Limit)
			},
		},
		{
			name:   "fail when repository returns error",
			userID:  uuid.New(),
			limit:  20,
			offset: 0,
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, userID uuid.UUID) {
				reader.On("FindByUserID", mock.Anything, userID, 20, 0).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get notifications",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := new(mocks.MockPortNotificationReader)
			tt.setupMocks(reader, tt.userID)

			useCase := &usecases.GetUserNotificationsUseCase{
				NotificationReader: reader,
			}

			ctx := tt.setupContext(context.Background(), tt.userID)
			result, err := useCase.Execute(ctx, tt.userID, tt.limit, tt.offset)

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

			reader.AssertExpectations(t)
		})
	}
}
