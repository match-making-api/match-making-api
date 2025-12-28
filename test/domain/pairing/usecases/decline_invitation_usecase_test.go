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

func TestDeclineInvitationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		invitationID  uuid.UUID
		userID        uuid.UUID
		setupMocks    func(*mocks.MockPortInvitationReader, *mocks.MockPortInvitationWriter, *mocks.MockInvitationNotifier, *pairing_entities.Invitation)
		expectedError string
	}{
		{
			name:         "successfully decline invitation",
			invitationID: uuid.New(),
			userID:       uuid.New(),
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				expirationDate := time.Now().Add(24 * time.Hour)
				inv.ExpirationDate = &expirationDate
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				writer.On("Save", mock.Anything, mock.MatchedBy(func(savedInv *pairing_entities.Invitation) bool {
					return savedInv.Status == pairing_entities.InvitationStatusDeclined && savedInv.DeclinedAt != nil
				})).Return(inv, nil)
				notifier.On("NotifyInvitationDeclined", mock.Anything, inv.ID, inv.UserID).Return(nil)
			},
		},
		{
			name:         "fail when invitation does not exist",
			invitationID: uuid.New(),
			userID:       uuid.New(),
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				reader.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("invitation not found"))
			},
			expectedError: "failed to get invitation",
		},
		{
			name:         "fail when invitation does not belong to user",
			invitationID: uuid.New(),
			userID:       uuid.New(),
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				inv.UserID = uuid.New() // Different user
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			expectedError: "does not belong to user",
		},
		{
			name:         "fail when invitation is expired",
			invitationID: uuid.New(),
			userID:       uuid.New(),
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				expirationDate := time.Now().Add(-1 * time.Hour)
				inv.ExpirationDate = &expirationDate
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			expectedError: "has expired",
		},
		{
			name:         "fail when invitation is not pending",
			invitationID: uuid.New(),
			userID:       uuid.New(),
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusDeclined
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
			},
			expectedError: "cannot be declined",
		},
		{
			name:         "fail when repository returns error on save",
			invitationID: uuid.New(),
			userID:       uuid.New(),
			setupMocks: func(reader *mocks.MockPortInvitationReader, writer *mocks.MockPortInvitationWriter, notifier *mocks.MockInvitationNotifier, inv *pairing_entities.Invitation) {
				inv.Status = pairing_entities.InvitationStatusPending
				expirationDate := time.Now().Add(24 * time.Hour)
				inv.ExpirationDate = &expirationDate
				reader.On("GetByID", mock.Anything, inv.ID).Return(inv, nil)
				writer.On("Save", mock.Anything, mock.AnythingOfType("*entities.Invitation")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to decline invitation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockPortInvitationReader)
			mockWriter := new(mocks.MockPortInvitationWriter)
			mockNotifier := new(mocks.MockInvitationNotifier)

			invitation := &pairing_entities.Invitation{
				BaseEntity: common.BaseEntity{ID: tt.invitationID},
				UserID:     tt.userID,
			}
			tt.setupMocks(mockReader, mockWriter, mockNotifier, invitation)

			useCase := &usecases.DeclineInvitationUseCase{
				InvitationReader: mockReader,
				InvitationWriter: mockWriter,
				Notifier:         mockNotifier,
			}

			ctx := context.Background()
			err := useCase.Execute(ctx, tt.invitationID, tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockReader.AssertExpectations(t)
			mockWriter.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}
