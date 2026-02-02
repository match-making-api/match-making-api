package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestSendBatchNotificationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		payload       usecases.SendBatchNotificationPayload
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortNotificationWriter, *mocks.MockPortUserNotificationPreferencesReader, *mocks.MockNotificationSender, []uuid.UUID)
		expectedError string
		expectedCount int
		expectedErrors int
	}{
		{
			name: "successfully send batch notifications",
			payload: usecases.SendBatchNotificationPayload{
				UserIDs:  []uuid.UUID{uuid.New(), uuid.New()},
				Channel:  pairing_entities.NotificationChannelInApp,
				Type:     pairing_entities.NotificationTypeSystemAnnouncement,
				Title:    "System Announcement",
				Message:  "System maintenance scheduled",
				Language: "en",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender, userIDs []uuid.UUID) {
				prefs := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: uuid.New()},
					uuid.New(),
					"en",
				)
				prefsReader.On("GetByUserID", mock.Anything, mock.Anything).Return(prefs, nil).Twice()
				sender.On("IsAvailable", mock.Anything).Return(true)
				sender.On("Send", mock.Anything, mock.Anything).Return(nil).Twice()
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					n.MarkAsSent()
				}).Return(mock.Anything, nil).Times(4) // 2 saves per notification (initial + after send)
			},
			expectedCount:  2,
			expectedErrors: 0,
		},
		{
			name: "fail when user is not admin",
			payload: usecases.SendBatchNotificationPayload{
				UserIDs: []uuid.UUID{uuid.New()},
				Channel: pairing_entities.NotificationChannelInApp,
				Type:    pairing_entities.NotificationTypeSystemAnnouncement,
				Title:   "System Announcement",
				Message: "System maintenance scheduled",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.UserAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender, userIDs []uuid.UUID) {
				// No mocks needed, should fail before calling them
			},
			expectedError: "only administrators can send batch notifications",
		},
		{
			name: "fail when user_ids is empty",
			payload: usecases.SendBatchNotificationPayload{
				UserIDs: []uuid.UUID{},
				Channel: pairing_entities.NotificationChannelInApp,
				Type:    pairing_entities.NotificationTypeSystemAnnouncement,
				Title:   "System Announcement",
				Message: "System maintenance scheduled",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender, userIDs []uuid.UUID) {
				// No mocks needed, should fail before calling them
			},
			expectedError: "user_ids list cannot be empty",
		},
		{
			name: "partial success when some users have disabled channels",
			payload: usecases.SendBatchNotificationPayload{
				UserIDs: []uuid.UUID{uuid.New(), uuid.New()},
				Channel: pairing_entities.NotificationChannelEmail,
				Type:    pairing_entities.NotificationTypeSystemAnnouncement,
				Title:   "System Announcement",
				Message: "System maintenance scheduled",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender, userIDs []uuid.UUID) {
				userID1 := userIDs[0]
				userID2 := userIDs[1]
				prefs1 := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: userID1},
					userID1,
					"en",
				)
				prefs2 := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: userID2},
					userID2,
					"en",
				)
				prefs2.DisabledChannels = []pairing_entities.NotificationChannel{pairing_entities.NotificationChannelEmail}
				prefsReader.On("GetByUserID", mock.Anything, userID1).Return(prefs1, nil)
				prefsReader.On("GetByUserID", mock.Anything, userID2).Return(prefs2, nil)
				sender.On("IsAvailable", mock.Anything).Return(true)
				sender.On("Send", mock.Anything, mock.Anything).Return(nil)
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					n.MarkAsSent()
				}).Return(mock.Anything, nil).Times(2) // Only one notification sent
			},
			expectedCount:  1,
			expectedErrors: 1,
		},
		{
			name: "fail when sender is not available",
			payload: usecases.SendBatchNotificationPayload{
				UserIDs: []uuid.UUID{uuid.New()},
				Channel: pairing_entities.NotificationChannelSMS,
				Type:    pairing_entities.NotificationTypeSystemAnnouncement,
				Title:   "System Announcement",
				Message: "System maintenance scheduled",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender, userIDs []uuid.UUID) {
				sender.On("IsAvailable", mock.Anything).Return(false)
			},
			expectedError: "sender for channel 2 is not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := new(mocks.MockPortNotificationWriter)
			prefsReader := new(mocks.MockPortUserNotificationPreferencesReader)
			sender := new(mocks.MockNotificationSender)

			tt.setupMocks(writer, prefsReader, sender, tt.payload.UserIDs)

			factory := usecases.NewNotificationSenderFactory()
			factory.RegisterSender(tt.payload.Channel, sender)

			useCase := &usecases.SendBatchNotificationUseCase{
				NotificationWriter:                writer,
				UserNotificationPreferencesReader: prefsReader,
				SenderFactory:                     factory,
			}

			ctx := tt.setupContext(context.Background())
			result, errs := useCase.Execute(ctx, tt.payload)

			if tt.expectedError != "" {
				assert.NotEmpty(t, errs)
				assert.Contains(t, errs[0].Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.Len(t, result, tt.expectedCount)
				assert.Len(t, errs, tt.expectedErrors)
			}

			writer.AssertExpectations(t)
			prefsReader.AssertExpectations(t)
			sender.AssertExpectations(t)
		})
	}
}
