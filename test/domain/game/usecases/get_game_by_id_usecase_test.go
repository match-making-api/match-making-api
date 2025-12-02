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
)

func TestGetGameByIDUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		gameID        uuid.UUID
		setupMocks    func(*MockGameReader)
		expectedError string
		validate      func(*testing.T, *game_entities.Game)
	}{
		{
			name:   "successfully get game by id",
			gameID: uuid.New(),
			setupMocks: func(reader *MockGameReader) {
				game := &game_entities.Game{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Game",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(game, nil)
			},
			validate: func(t *testing.T, game *game_entities.Game) {
				assert.NotNil(t, game)
				assert.Equal(t, "Test Game", game.Name)
			},
		},
		{
			name:   "fail when game not found",
			gameID: uuid.New(),
			setupMocks: func(reader *MockGameReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "failed to get game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(MockGameReader)
			tt.setupMocks(mockReader)

			useCase := usecases.NewGetGameByIDUseCase(mockReader)

			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.gameID)

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
