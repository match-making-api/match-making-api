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

func TestGetGameModeByIDUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		gameModeID    uuid.UUID
		setupMocks    func(*mocks.MockGameModeReader)
		expectedError string
		validate      func(*testing.T, *game_entities.GameMode)
	}{
		{
			name:       "successfully get game mode by id",
			gameModeID: uuid.New(),
			setupMocks: func(reader *mocks.MockGameModeReader) {
				gameMode := &game_entities.GameMode{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Game Mode",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(gameMode, nil)
			},
			validate: func(t *testing.T, gameMode *game_entities.GameMode) {
				assert.NotNil(t, gameMode)
				assert.Equal(t, "Test Game Mode", gameMode.Name)
			},
		},
		{
			name:       "fail when game mode not found",
			gameModeID: uuid.New(),
			setupMocks: func(reader *mocks.MockGameModeReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "failed to get game mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockGameModeReader)
			tt.setupMocks(mockReader)

			useCase := usecases.NewGetGameModeByIDUseCase(mockReader)

			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.gameModeID)

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

			mockReader.AssertExpectations(t)
		})
	}
}
