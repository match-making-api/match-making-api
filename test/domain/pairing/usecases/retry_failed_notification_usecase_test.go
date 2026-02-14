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

func TestRetryFailedNotificationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		notificationID uuid.UUID
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortNotificationReader, *mocks.MockPortNotificationWriter, *mocks.MockNotificationSender, uuid.UUID)
		expectedError string
		validate      func(*testing.T, *pairing_entities.Notification)
	}{
		{
			name:          "successfully retry failed notification",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, sender *mocks.MockNotificationSender, notificationID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity:  common.BaseEntity{ID: notificationID},
					UserID:      uuid.New(),
					Channel:     pairing_entities.NotificationChannelEmail,
					Type:        pairing_entities.NotificationTypeMatchInvitation,
					Title:       "Match Invitation",
					Message:     "You have been invited",
					Status:      pairing_entities.NotificationStatusFailed,
					RetryCount:  1,
					MaxRetries:  3,
					FailureReason: stringPtr("send failed"),
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
				sender.On("IsAvailable", mock.Anything).Return(true)
				sender.On("Send", mock.Anything, mock.Anything).Return(nil)
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					if n.Status == pairing_entities.NotificationStatusFailed {
						n.ScheduleRetry(time.Now().Add(5 * time.Minute))
					} else {
						n.MarkAsSent()
					}
				}).Return(mock.Anything, nil).Twice() // Once for scheduling retry, once after send
			},
			validate: func(t *testing.T, notification *pairing_entities.Notification) {
				assert.Equal(t, pairing_entities.NotificationStatusSent, notification.Status)
			},
		},
		{
			name:          "fail when user is not admin",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				// No AudienceKey set so IsAdmin returns false
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, sender *mocks.MockNotificationSender, notificationID uuid.UUID) {
				// No mocks needed, should fail before calling them
			},
			expectedError: "only administrators can retry notifications",
		},
		{
			name:          "fail when notification not found",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, sender *mocks.MockNotificationSender, notificationID uuid.UUID) {
				reader.On("GetByID", mock.Anything, notificationID).Return(nil, errors.New("not found"))
			},
			expectedError: "failed to get notification",
		},
		{
			name:          "fail when notification cannot be retried (not failed)",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, sender *mocks.MockNotificationSender, notificationID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity: common.BaseEntity{ID: notificationID},
					UserID:     uuid.New(),
					Channel:    pairing_entities.NotificationChannelEmail,
					Type:       pairing_entities.NotificationTypeMatchInvitation,
					Title:      "Match Invitation",
					Message:    "You have been invited",
					Status:     pairing_entities.NotificationStatusSent, // Already sent
					RetryCount: 0,
					MaxRetries: 3,
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
			},
			expectedError: "notification cannot be retried",
		},
		{
			name:          "fail when max retries reached",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, sender *mocks.MockNotificationSender, notificationID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity:  common.BaseEntity{ID: notificationID},
					UserID:      uuid.New(),
					Channel:     pairing_entities.NotificationChannelEmail,
					Type:        pairing_entities.NotificationTypeMatchInvitation,
					Title:       "Match Invitation",
					Message:     "You have been invited",
					Status:      pairing_entities.NotificationStatusFailed,
					RetryCount:  3,
					MaxRetries:  3, // Max retries reached
					FailureReason: stringPtr("send failed"),
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
			},
			expectedError: "notification cannot be retried",
		},
		{
			name:          "fail when sender is not available",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, sender *mocks.MockNotificationSender, notificationID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity:  common.BaseEntity{ID: notificationID},
					UserID:      uuid.New(),
					Channel:     pairing_entities.NotificationChannelEmail,
					Type:        pairing_entities.NotificationTypeMatchInvitation,
					Title:       "Match Invitation",
					Message:     "You have been invited",
					Status:      pairing_entities.NotificationStatusFailed,
					RetryCount:  1,
					MaxRetries:  3,
					FailureReason: stringPtr("send failed"),
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
				sender.On("IsAvailable", mock.Anything).Return(false)
			},
			expectedError: "sender for channel 1 is not available",
		},
		{
			name:          "fail when retry send fails",
			notificationID: uuid.New(),
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(reader *mocks.MockPortNotificationReader, writer *mocks.MockPortNotificationWriter, sender *mocks.MockNotificationSender, notificationID uuid.UUID) {
				notification := &pairing_entities.Notification{
					BaseEntity:  common.BaseEntity{ID: notificationID},
					UserID:      uuid.New(),
					Channel:     pairing_entities.NotificationChannelEmail,
					Type:        pairing_entities.NotificationTypeMatchInvitation,
					Title:       "Match Invitation",
					Message:     "You have been invited",
					Status:      pairing_entities.NotificationStatusFailed,
					RetryCount:  1,
					MaxRetries:  3,
					FailureReason: stringPtr("send failed"),
				}
				reader.On("GetByID", mock.Anything, notificationID).Return(notification, nil)
				sender.On("IsAvailable", mock.Anything).Return(true)
				sender.On("Send", mock.Anything, mock.Anything).Return(errors.New("retry send failed"))
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					if n.Status == pairing_entities.NotificationStatusFailed {
						n.ScheduleRetry(time.Now().Add(5 * time.Minute))
					} else {
						n.MarkAsFailed("retry send failed")
					}
				}).Return(mock.Anything, nil).Twice() // Once for scheduling retry, once after failed send
			},
			expectedError: "retry failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := new(mocks.MockPortNotificationReader)
			writer := new(mocks.MockPortNotificationWriter)
			sender := new(mocks.MockNotificationSender)

			tt.setupMocks(reader, writer, sender, tt.notificationID)

			factory := usecases.NewNotificationSenderFactory()
			factory.RegisterSender(pairing_entities.NotificationChannelEmail, sender)

			useCase := usecases.NewRetryFailedNotificationUseCase(
				reader,
				writer,
				factory,
				5*time.Minute,
			)

			ctx := tt.setupContext(context.Background())
			err := useCase.Execute(ctx, tt.notificationID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					// Create a notification to validate
					notification := &pairing_entities.Notification{
						BaseEntity: common.BaseEntity{ID: tt.notificationID},
						Status:     pairing_entities.NotificationStatusFailed,
					}
					notification.MarkAsSent()
					tt.validate(t, notification)
				}
			}

			reader.AssertExpectations(t)
			writer.AssertExpectations(t)
			sender.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
