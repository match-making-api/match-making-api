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

// MockExternalInvitationNotifier is a mock implementation of usecases.ExternalInvitationNotifier using testify/mock
type MockExternalInvitationNotifier struct {
	mock.Mock
}

// Ensure MockExternalInvitationNotifier implements usecases.ExternalInvitationNotifier
var _ usecases.ExternalInvitationNotifier = (*MockExternalInvitationNotifier)(nil)

func (m *MockExternalInvitationNotifier) NotifyInvitationCreated(ctx context.Context, invitationID uuid.UUID, email string, fullName string, message string, registrationToken string, matchID *uuid.UUID, eventID *uuid.UUID) error {
	args := m.Called(ctx, invitationID, email, fullName, message, registrationToken, matchID, eventID)
	return args.Error(0)
}

func (m *MockExternalInvitationNotifier) NotifyInvitationAccepted(ctx context.Context, invitationID uuid.UUID, email string, fullName string, userID uuid.UUID) error {
	args := m.Called(ctx, invitationID, email, fullName, userID)
	return args.Error(0)
}

func (m *MockExternalInvitationNotifier) NotifyInvitationExpired(ctx context.Context, invitationID uuid.UUID, email string, fullName string) error {
	args := m.Called(ctx, invitationID, email, fullName)
	return args.Error(0)
}
