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

func TestCreateGameModeUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		gameMode      *game_entities.GameMode
		setupMocks    func(*mocks.MockPortGameModeWriter, *mocks.MockPortGameModeReader)
		expectedError string
		validate      func(*testing.T, *game_entities.GameMode)
	}{
		{
			name: "successfully create game mode",
			gameMode: &game_entities.GameMode{
				GameID:      uuid.FromStringOrNil(google_uuid.New().String()),
				Name:        "Test Game Mode",
				Description: "Test Description",
			},
			setupMocks: func(writer *mocks.MockPortGameModeWriter, reader *mocks.MockPortGameModeReader) {
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.GameMode{}, nil)
				writer.On("Create", mock.Anything, mock.AnythingOfType("*entities.GameMode")).Return(&game_entities.GameMode{
					BaseEntity:  common.BaseEntity{ID: google_uuid.New()},
					GameID:      uuid.Must(uuid.NewV4()),
					Name:        "Test Game Mode",
					Description: "Test Description",
				}, nil)
			},
			validate: func(t *testing.T, gameMode *game_entities.GameMode) {
				assert.NotEqual(t, google_uuid.Nil, gameMode.BaseEntity.ID)
				assert.Equal(t, "Test Game Mode", gameMode.Name)
				assert.NotEqual(t, uuid.Nil, gameMode.GameID)
			},
		},
		{
			name: "fail when game mode name is empty",
			gameMode: &game_entities.GameMode{
				GameID: uuid.FromStringOrNil(google_uuid.New().String()),
				Name:   "",
			},
			setupMocks: func(writer *mocks.MockPortGameModeWriter, reader *mocks.MockPortGameModeReader) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "game mode name is required",
		},
		{
			name: "fail when game_id is missing",
			gameMode: &game_entities.GameMode{
				GameID: uuid.Nil,
				Name:   "Test Game Mode",
			},
			setupMocks: func(writer *mocks.MockPortGameModeWriter, reader *mocks.MockPortGameModeReader) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "game_id is required",
		},
		{
			name: "fail when duplicate game mode name exists for same game",
			gameMode: &game_entities.GameMode{
				GameID: uuid.FromStringOrNil("12345678-1234-1234-1234-123456789012"),
				Name:   "Existing Game Mode",
			},
			setupMocks: func(writer *mocks.MockPortGameModeWriter, reader *mocks.MockPortGameModeReader) {
				existingGameMode := &game_entities.GameMode{
					BaseEntity: common.BaseEntity{ID: google_uuid.New()},
					GameID:     uuid.FromStringOrNil("12345678-1234-1234-1234-123456789012"),
					Name:       "Existing Game Mode",
				}
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.GameMode{existingGameMode}, nil)
			},
			expectedError: "game mode with name 'Existing Game Mode' already exists for this game",
		},
		{
			name: "fail when repository returns error",
			gameMode: &game_entities.GameMode{
				GameID: uuid.FromStringOrNil(google_uuid.New().String()),
				Name:   "Test Game Mode",
			},
			setupMocks: func(writer *mocks.MockPortGameModeWriter, reader *mocks.MockPortGameModeReader) {
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.GameMode{}, nil)
				writer.On("Create", mock.Anything, mock.AnythingOfType("*entities.GameMode")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to create game mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(mocks.MockPortGameModeWriter)
			mockReader := new(mocks.MockPortGameModeReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewCreateGameModeUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, google_uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, google_uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, google_uuid.New())

			result, err := useCase.Execute(ctx, tt.gameMode)

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
