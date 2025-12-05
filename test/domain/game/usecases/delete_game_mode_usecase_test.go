package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/usecases"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestDeleteGameModeUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		gameModeID    uuid.UUID
		setupMocks    func(*mocks.MockGameModeWriter, *mocks.MockGameModeReader)
		expectedError string
	}{
		{
			name:       "successfully delete game mode",
			gameModeID: uuid.New(),
			setupMocks: func(writer *mocks.MockGameModeWriter, reader *mocks.MockGameModeReader) {
				existingGameMode := &game_entities.GameMode{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Game Mode",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGameMode, nil)
				writer.On("Delete", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:       "fail when game mode not found",
			gameModeID: uuid.New(),
			setupMocks: func(writer *mocks.MockGameModeWriter, reader *mocks.MockGameModeReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "game mode not found",
		},
		{
			name:       "fail when delete fails",
			gameModeID: uuid.New(),
			setupMocks: func(writer *mocks.MockGameModeWriter, reader *mocks.MockGameModeReader) {
				existingGameMode := &game_entities.GameMode{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Game Mode",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGameMode, nil)
				writer.On("Delete", mock.Anything, mock.Anything).Return(errors.New("delete failed"))
			},
			expectedError: "failed to delete game mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(mocks.MockGameModeWriter)
			mockReader := new(mocks.MockGameModeReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewDeleteGameModeUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

			err := useCase.Execute(ctx, tt.gameModeID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockWriter.AssertExpectations(t)
			mockReader.AssertExpectations(t)
		})
	}
}
