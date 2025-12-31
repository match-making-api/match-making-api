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

func TestSendNotificationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		payload       usecases.SendNotificationPayload
		setupContext  func(context.Context) context.Context
		setupMocks    func(*mocks.MockPortNotificationWriter, *mocks.MockPortNotificationReader, *mocks.MockPortUserNotificationPreferencesReader, *mocks.MockNotificationSender)
		expectedError string
		validate      func(*testing.T, *pairing_entities.Notification)
	}{
		{
			name: "successfully send in-app notification",
			payload: usecases.SendNotificationPayload{
				UserID:   uuid.New(),
				Channel:  pairing_entities.NotificationChannelInApp,
				Type:     pairing_entities.NotificationTypeMatchInvitation,
				Title:    "Match Invitation",
				Message:  "You have been invited to a match",
				Metadata: map[string]interface{}{"match_id": uuid.New().String()},
				Language: "en",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, reader *mocks.MockPortNotificationReader, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender) {
				userID := uuid.New()
				prefs := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: userID},
					userID,
					"en",
				)
				prefsReader.On("GetByUserID", mock.Anything, mock.Anything).Return(prefs, nil)
				sender.On("IsAvailable", mock.Anything).Return(true)
				sender.On("Send", mock.Anything, mock.Anything).Return(nil)
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					n.MarkAsSent()
				}).Return(mock.Anything, nil).Twice()
			},
			validate: func(t *testing.T, notification *pairing_entities.Notification) {
				assert.NotNil(t, notification)
				assert.Equal(t, pairing_entities.NotificationChannelInApp, notification.Channel)
				assert.Equal(t, pairing_entities.NotificationTypeMatchInvitation, notification.Type)
				assert.Equal(t, "Match Invitation", notification.Title)
				assert.Equal(t, pairing_entities.NotificationStatusSent, notification.Status)
			},
		},
		{
			name: "fail when channel is disabled",
			payload: usecases.SendNotificationPayload{
				UserID:  uuid.New(),
				Channel: pairing_entities.NotificationChannelEmail,
				Type:    pairing_entities.NotificationTypeMatchInvitation,
				Title:   "Match Invitation",
				Message: "You have been invited to a match",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, reader *mocks.MockPortNotificationReader, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender) {
				userID := uuid.New()
				prefs := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: userID},
					userID,
					"en",
				)
				prefs.DisabledChannels = []pairing_entities.NotificationChannel{pairing_entities.NotificationChannelEmail}
				prefsReader.On("GetByUserID", mock.Anything, mock.Anything).Return(prefs, nil)
			},
			expectedError: "channel 1 is disabled for user",
		},
		{
			name: "fail when notification type is disabled",
			payload: usecases.SendNotificationPayload{
				UserID:  uuid.New(),
				Channel: pairing_entities.NotificationChannelInApp,
				Type:    pairing_entities.NotificationTypeMatchInvitation,
				Title:   "Match Invitation",
				Message: "You have been invited to a match",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, reader *mocks.MockPortNotificationReader, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender) {
				userID := uuid.New()
				prefs := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: userID},
					userID,
					"en",
				)
				prefs.TypePreferences[pairing_entities.NotificationTypeMatchInvitation] = false
				prefsReader.On("GetByUserID", mock.Anything, mock.Anything).Return(prefs, nil)
			},
			expectedError: "notification type 0 is disabled for user",
		},
		{
			name: "fail during do not disturb time",
			payload: usecases.SendNotificationPayload{
				UserID:  uuid.New(),
				Channel: pairing_entities.NotificationChannelInApp,
				Type:    pairing_entities.NotificationTypeMatchInvitation,
				Title:   "Match Invitation",
				Message: "You have been invited to a match",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, reader *mocks.MockPortNotificationReader, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender) {
				userID := uuid.New()
				prefs := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: userID},
					userID,
					"en",
				)
				now := time.Now()
				start := now.Format("15:04")
				end := now.Add(1 * time.Hour).Format("15:04")
				prefs.DoNotDisturbStart = &start
				prefs.DoNotDisturbEnd = &end
				prefsReader.On("GetByUserID", mock.Anything, mock.Anything).Return(prefs, nil)
			},
			expectedError: "cannot send notification during do not disturb time",
		},
		{
			name: "fail when sender is not available",
			payload: usecases.SendNotificationPayload{
				UserID:  uuid.New(),
				Channel: pairing_entities.NotificationChannelEmail,
				Type:    pairing_entities.NotificationTypeMatchInvitation,
				Title:   "Match Invitation",
				Message: "You have been invited to a match",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, reader *mocks.MockPortNotificationReader, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender) {
				userID := uuid.New()
				prefs := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: userID},
					userID,
					"en",
				)
				prefsReader.On("GetByUserID", mock.Anything, mock.Anything).Return(prefs, nil)
				sender.On("GetChannel").Return(pairing_entities.NotificationChannelEmail)
				sender.On("IsAvailable", mock.Anything).Return(false)
			},
			expectedError: "sender for channel 1 is not available",
		},
		{
			name: "fail when sender fails to send",
			payload: usecases.SendNotificationPayload{
				UserID:  uuid.New(),
				Channel: pairing_entities.NotificationChannelInApp,
				Type:    pairing_entities.NotificationTypeMatchInvitation,
				Title:   "Match Invitation",
				Message: "You have been invited to a match",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, reader *mocks.MockPortNotificationReader, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender) {
				userID := uuid.New()
				prefs := pairing_entities.NewUserNotificationPreferences(
					common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New(), UserID: userID},
					userID,
					"en",
				)
				prefsReader.On("GetByUserID", mock.Anything, mock.Anything).Return(prefs, nil)
				sender.On("IsAvailable", mock.Anything).Return(true)
				sender.On("Send", mock.Anything, mock.Anything).Return(errors.New("send failed"))
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					n.MarkAsSent()
				}).Return(mock.Anything, nil).Twice()
			},
			expectedError: "failed to send notification",
		},
		{
			name: "use default preferences when preferences not found",
			payload: usecases.SendNotificationPayload{
				UserID:  uuid.New(),
				Channel: pairing_entities.NotificationChannelInApp,
				Type:    pairing_entities.NotificationTypeMatchInvitation,
				Title:   "Match Invitation",
				Message: "You have been invited to a match",
			},
			setupContext: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
				ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)
				return ctx
			},
			setupMocks: func(writer *mocks.MockPortNotificationWriter, reader *mocks.MockPortNotificationReader, prefsReader *mocks.MockPortUserNotificationPreferencesReader, sender *mocks.MockNotificationSender) {
				prefsReader.On("GetByUserID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
				sender.On("GetChannel").Return(pairing_entities.NotificationChannelInApp)
				sender.On("IsAvailable", mock.Anything).Return(true)
				sender.On("Send", mock.Anything, mock.Anything).Return(nil)
				writer.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					n := args.Get(1).(*pairing_entities.Notification)
					n.MarkAsSent()
				}).Return(mock.Anything, nil).Twice()
			},
			validate: func(t *testing.T, notification *pairing_entities.Notification) {
				assert.NotNil(t, notification)
				assert.Equal(t, pairing_entities.NotificationStatusSent, notification.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := new(mocks.MockPortNotificationWriter)
			reader := new(mocks.MockPortNotificationReader)
			prefsReader := new(mocks.MockPortUserNotificationPreferencesReader)
			sender := new(mocks.MockNotificationSender)

			tt.setupMocks(writer, reader, prefsReader, sender)

			factory := usecases.NewNotificationSenderFactory()
			factory.RegisterSender(tt.payload.Channel, sender)

			useCase := &usecases.SendNotificationUseCase{
				NotificationWriter:                writer,
				NotificationReader:                reader,
				UserNotificationPreferencesReader: prefsReader,
				SenderFactory:                     factory,
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

			writer.AssertExpectations(t)
			reader.AssertExpectations(t)
			prefsReader.AssertExpectations(t)
			sender.AssertExpectations(t)
		})
	}
}
