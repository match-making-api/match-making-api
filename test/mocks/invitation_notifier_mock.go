package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
)

// MockInvitationNotifier is a mock implementation of usecases.InvitationNotifier using testify/mock
type MockInvitationNotifier struct {
	mock.Mock
}

// Ensure MockInvitationNotifier implements usecases.InvitationNotifier
var _ usecases.InvitationNotifier = (*MockInvitationNotifier)(nil)

func (m *MockInvitationNotifier) NotifyInvitationCreated(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID, message string) error {
	args := m.Called(ctx, invitationID, userID, message)
	return args.Error(0)
}

func (m *MockInvitationNotifier) NotifyInvitationAccepted(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, invitationID, userID)
	return args.Error(0)
}

func (m *MockInvitationNotifier) NotifyInvitationDeclined(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, invitationID, userID)
	return args.Error(0)
}

func (m *MockInvitationNotifier) NotifyInvitationRevoked(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, invitationID, userID)
	return args.Error(0)
}
