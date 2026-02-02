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

func TestMarkNotificationReadUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		notificationID uuid.UUID
		setupContext  func(context.Context, uuid.UUID) context.Context
		setupMocks    func(*mocks.MockPortNotificationReader, *mocks.MockPortNotificationWriter, uuid.UUID, uuid.UUID)
		expectedError string
		validate      func(*testing.T, *pairing_entities.Notification)
	}{
		{
			name:          "successfully mark in-app notification as read",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, notificationID uuid.UUID, userID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity: common.BaseEntity{ID: notificationID},
					UserID:     userID,
					Channel:    pairing_entities.NotificationChannelInApp,
					Type:       pairing_entities.NotificationTypeMatchInvitation,
					Title:      "Match Invitation",
					Message:    "You have been invited",
					Status:     pairing_entities.NotificationStatusSent,
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					n.MarkAsRead()
				}).Return(mock.Anything, nil)
			},
			validate: func(t *testing.T, notification *pairing_entities.Notification) {
				assert.NotNil(t, notification.ReadAt)
			},
		},
		{
			name:          "fail when notification not found",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, notificationID uuid.UUID, userID uuid.UUID) {
				reader.On("GetByID", mock.Anything, notificationID).Return(nil, errors.New("not found"))
			},
			expectedError: "failed to get notification",
		},
		{
			name:          "fail when user tries to mark other user's notification as read",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context, _ uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New()) // Different user
				ctx = context.WithValue(ctx, common.AudienceKey, common.UserAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, notificationID uuid.UUID, userID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity: common.BaseEntity{ID: notificationID},
					UserID:     userID,
					Channel:    pairing_entities.NotificationChannelInApp,
					Type:       pairing_entities.NotificationTypeMatchInvitation,
					Title:      "Match Invitation",
					Message:    "You have been invited",
					Status:     pairing_entities.NotificationStatusSent,
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
			},
			expectedError: "user can only mark their own notifications as read",
		},
		{
			name:          "admin can mark any notification as read",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context, _ uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, notificationID uuid.UUID, userID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity: common.BaseEntity{ID: notificationID},
					UserID:     userID,
					Channel:    pairing_entities.NotificationChannelInApp,
					Type:       pairing_entities.NotificationTypeMatchInvitation,
					Title:      "Match Invitation",
					Message:    "You have been invited",
					Status:     pairing_entities.NotificationStatusSent,
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					n.MarkAsRead()
				}).Return(mock.Anything, nil)
			},
			validate: func(t *testing.T, notification *pairing_entities.Notification) {
				assert.NotNil(t, notification.ReadAt)
			},
		},
		{
			name:          "fail when notification is not in-app",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, notificationID uuid.UUID, userID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity: common.BaseEntity{ID: notificationID},
					UserID:     userID,
					Channel:    pairing_entities.NotificationChannelEmail, // Not in-app
					Type:       pairing_entities.NotificationTypeMatchInvitation,
					Title:      "Match Invitation",
					Message:    "You have been invited",
					Status:     pairing_entities.NotificationStatusSent,
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
			},
			expectedError: "only in-app notifications can be marked as read",
		},
		{
			name:          "fail when save fails",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context, userID uuid.UUID) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, notificationID uuid.UUID, userID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity: common.BaseEntity{ID: notificationID},
					UserID:     userID,
					Channel:    pairing_entities.NotificationChannelInApp,
					Type:       pairing_entities.NotificationTypeMatchInvitation,
					Title:      "Match Invitation",
					Message:    "You have been invited",
					Status:     pairing_entities.NotificationStatusSent,
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
				writer.On("Save", mock.Anything, mock.Anything).Return(nil, errors.New("save failed"))
			},
			expectedError: "failed to update notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := new(mocks.MockPortNotificationReader)
			writer := new(mocks.MockPortNotificationWriter)
			userID := uuid.New()
			tt.setupMocks(reader, writer, tt.notificationID, userID)

			useCase := &usecases.MarkNotificationReadUseCase{
				NotificationReader: reader,
				NotificationWriter: writer,
			}

			ctx := tt.setupContext(context.Background(), userID)
			err := useCase.Execute(ctx, tt.notificationID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					// Get the notification from the mock to validate
					notification := &pairing_entities.Notification{
						BaseEntity: common.BaseEntity{ID: tt.notificationID},
						UserID:     userID,
						Channel:    pairing_entities.NotificationChannelInApp,
					}
					notification.MarkAsRead()
					tt.validate(t, notification)
				}
			}

			reader.AssertExpectations(t)
			writer.AssertExpectations(t)
		})
	}
}
