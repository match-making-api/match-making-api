package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
)

// MockNotificationSender is a mock implementation of usecases.NotificationSender using testify/mock
type MockNotificationSender struct {
	mock.Mock
}

// Ensure MockNotificationSender implements usecases.NotificationSender
var _ usecases.NotificationSender = (*MockNotificationSender)(nil)

func (m *MockNotificationSender) Send(ctx context.Context, notification *pairing_entities.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationSender) GetChannel() pairing_entities.NotificationChannel {
	args := m.Called()
	return args.Get(0).(pairing_entities.NotificationChannel)
}

func (m *MockNotificationSender) IsAvailable(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}
