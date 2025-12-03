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

func TestDeleteGameUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		gameID        uuid.UUID
		existingGame  *game_entities.Game
		setupMocks    func(*mocks.MockGameWriter, *mocks.MockGameReader)
		expectedError string
	}{
		{
			name:   "successfully disable enabled game",
			gameID: uuid.New(),
			existingGame: &game_entities.Game{
				BaseEntity: common.BaseEntity{ID: uuid.New()},
				Name:       "Test Game",
				Enabled:    true,
			},
			setupMocks: func(writer *mocks.MockGameWriter, reader *mocks.MockGameReader) {
				existingGame := &game_entities.Game{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Game",
					Enabled:    true,
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGame, nil)
				disabledGame := *existingGame
				disabledGame.Enabled = false
				writer.On("Update", mock.Anything, mock.AnythingOfType("*entities.Game")).Return(&disabledGame, nil)
			},
		},
		{
			name:   "successfully delete disabled game",
			gameID: uuid.New(),
			existingGame: &game_entities.Game{
				BaseEntity: common.BaseEntity{ID: uuid.New()},
				Name:       "Test Game",
				Enabled:    false,
			},
			setupMocks: func(writer *mocks.MockGameWriter, reader *mocks.MockGameReader) {
				existingGame := &game_entities.Game{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Game",
					Enabled:    false,
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGame, nil)
				writer.On("Delete", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:   "fail when game not found",
			gameID: uuid.New(),
			setupMocks: func(writer *mocks.MockGameWriter, reader *mocks.MockGameReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "game not found",
		},
		{
			name:   "fail when disable update fails",
			gameID: uuid.New(),
			existingGame: &game_entities.Game{
				BaseEntity: common.BaseEntity{ID: uuid.New()},
				Name:       "Test Game",
				Enabled:    true,
			},
			setupMocks: func(writer *mocks.MockGameWriter, reader *mocks.MockGameReader) {
				existingGame := &game_entities.Game{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Game",
					Enabled:    true,
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGame, nil)
				writer.On("Update", mock.Anything, mock.AnythingOfType("*entities.Game")).Return(nil, errors.New("update failed"))
			},
			expectedError: "failed to disable game",
		},
		{
			name:   "fail when delete fails",
			gameID: uuid.New(),
			existingGame: &game_entities.Game{
				BaseEntity: common.BaseEntity{ID: uuid.New()},
				Name:       "Test Game",
				Enabled:    false,
			},
			setupMocks: func(writer *mocks.MockGameWriter, reader *mocks.MockGameReader) {
				existingGame := &game_entities.Game{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Game",
					Enabled:    false,
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGame, nil)
				writer.On("Delete", mock.Anything, mock.Anything).Return(errors.New("delete failed"))
			},
			expectedError: "failed to delete game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(mocks.MockGameWriter)
			mockReader := new(mocks.MockGameReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewDeleteGameUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

			err := useCase.Execute(ctx, tt.gameID)

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
