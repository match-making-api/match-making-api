package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	google_uuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/usecases"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestUpdateGameModeUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		gameModeID    google_uuid.UUID
		gameMode      *game_entities.GameMode
		setupMocks    func(*mocks.MockPortGameModeWriter, *mocks.MockPortGameModeReader)
		expectedError string
		validate      func(*testing.T, *game_entities.GameMode)
	}{
		{
			name:       "successfully update game mode",
			gameModeID: google_uuid.New(),
			gameMode: &game_entities.GameMode{
				GameID:      uuid.FromStringOrNil(google_uuid.New().String()),
				Name:        "Updated Game Mode",
				Description: "Updated Description",
			},
			setupMocks: func(writer *mocks.MockPortGameModeWriter, reader *mocks.MockPortGameModeReader) {
				existingGameMode := &game_entities.GameMode{
					BaseEntity:  common.BaseEntity{ID: google_uuid.New()},
					GameID:      uuid.FromStringOrNil(google_uuid.New().String()),
					Name:        "Original Game Mode",
					Description: "Original Description",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGameMode, nil)
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.GameMode{}, nil)
				writer.On("Update", mock.Anything, mock.AnythingOfType("*entities.GameMode")).Return(existingGameMode, nil)
			},
			validate: func(t *testing.T, gameMode *game_entities.GameMode) {
				assert.Equal(t, "Updated Game Mode", gameMode.Name)
				assert.Equal(t, "Updated Description", gameMode.Description)
			},
		},
		{
			name:       "fail when game mode not found",
			gameModeID: google_uuid.New(),
			gameMode: &game_entities.GameMode{
				GameID: uuid.FromStringOrNil(google_uuid.New().String()),
				Name:   "Test Game Mode",
			},
			setupMocks: func(writer *mocks.MockPortGameModeWriter, reader *mocks.MockPortGameModeReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "game mode not found",
		},
		{
			name:       "fail when duplicate name exists for same game",
			gameModeID: google_uuid.New(),
			gameMode: &game_entities.GameMode{
				GameID: uuid.FromStringOrNil(google_uuid.New().String()),
				Name:   "Duplicate Name",
			},
			setupMocks: func(writer *mocks.MockPortGameModeWriter, reader *mocks.MockPortGameModeReader) {
				existingGameMode := &game_entities.GameMode{
					BaseEntity: common.BaseEntity{ID: google_uuid.New()},
					GameID:     uuid.FromStringOrNil(google_uuid.New().String()),
					Name:       "Original Game Mode",
				}
				duplicateGameMode := &game_entities.GameMode{
					BaseEntity: common.BaseEntity{ID: google_uuid.New()},
					GameID:     uuid.FromStringOrNil(google_uuid.New().String()),
					Name:       "Duplicate Name",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGameMode, nil)
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.GameMode{existingGameMode, duplicateGameMode}, nil)
			},
			expectedError: "game mode with name 'Duplicate Name' already exists for this game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(mocks.MockPortGameModeWriter)
			mockReader := new(mocks.MockPortGameModeReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewUpdateGameModeUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, google_uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, google_uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, google_uuid.New())

			result, err := useCase.Execute(ctx, tt.gameModeID, tt.gameMode)

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

			mockWriter.AssertExpectations(t)
			mockReader.AssertExpectations(t)
		})
	}
}
