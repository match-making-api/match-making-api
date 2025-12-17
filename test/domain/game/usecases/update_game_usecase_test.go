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
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/usecases"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestUpdateGameUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		gameID        uuid.UUID
		game          *game_entities.Game
		setupMocks    func(*mocks.MockPortGameWriter, *mocks.MockPortGameReader)
		expectedError string
		validate      func(*testing.T, *game_entities.Game)
	}{
		{
			name:   "successfully update game",
			gameID: uuid.New(),
			game: &game_entities.Game{
				Name:              "Updated Game",
				Description:       "Updated Description",
				MinPlayersPerTeam: 2,
				MaxPlayersPerTeam: 6,
				NumberOfTeams:     2,
				MaxDuration:       45 * time.Minute,
			},
			setupMocks: func(writer *mocks.MockPortGameWriter, reader *mocks.MockPortGameReader) {
				existingGame := &game_entities.Game{
					BaseEntity:        common.BaseEntity{ID: uuid.New()},
					Name:              "Original Game",
					Description:       "Original Description",
					MinPlayersPerTeam: 1,
					MaxPlayersPerTeam: 5,
					NumberOfTeams:     2,
					MaxDuration:       30 * time.Minute,
					Enabled:           true,
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGame, nil)
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Game{}, nil)
				writer.On("Update", mock.Anything, mock.AnythingOfType("*entities.Game")).Return(existingGame, nil)
			},
			validate: func(t *testing.T, game *game_entities.Game) {
				assert.Equal(t, "Updated Game", game.Name)
				assert.Equal(t, "Updated Description", game.Description)
				assert.Equal(t, 2, game.MinPlayersPerTeam)
				assert.Equal(t, 6, game.MaxPlayersPerTeam)
			},
		},
		{
			name:   "fail when game not found",
			gameID: uuid.New(),
			game: &game_entities.Game{
				Name:              "Test Game",
				MinPlayersPerTeam: 1,
				MaxPlayersPerTeam: 5,
				NumberOfTeams:     2,
			},
			setupMocks: func(writer *mocks.MockPortGameWriter, reader *mocks.MockPortGameReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "game not found",
		},
		{
			name:   "fail when duplicate name exists",
			gameID: uuid.New(),
			game: &game_entities.Game{
				Name:              "Duplicate Name",
				MinPlayersPerTeam: 1,
				MaxPlayersPerTeam: 5,
				NumberOfTeams:     2,
			},
			setupMocks: func(writer *mocks.MockPortGameWriter, reader *mocks.MockPortGameReader) {
				existingGame := &game_entities.Game{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Original Game",
				}
				duplicateGame := &game_entities.Game{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Duplicate Name",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingGame, nil)
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Game{existingGame, duplicateGame}, nil)
			},
			expectedError: "game with name 'Duplicate Name' already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(mocks.MockPortGameWriter)
			mockReader := new(mocks.MockPortGameReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewUpdateGameUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

			result, err := useCase.Execute(ctx, tt.gameID, tt.game)

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
