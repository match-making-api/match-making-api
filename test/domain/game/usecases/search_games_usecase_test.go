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

func TestSearchGamesUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockGameReader)
		expectedError string
		validate      func(*testing.T, []*game_entities.Game)
	}{
		{
			name: "successfully search games",
			setupMocks: func(reader *MockGameReader) {
				games := []*game_entities.Game{
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Name:       "Game 1",
					},
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Name:       "Game 2",
					},
				}
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return(games, nil)
			},
			validate: func(t *testing.T, games []*game_entities.Game) {
				assert.NotNil(t, games)
				assert.Len(t, games, 2)
				assert.Equal(t, "Game 1", games[0].Name)
				assert.Equal(t, "Game 2", games[1].Name)
			},
		},
		{
			name: "return empty list when no games found",
			setupMocks: func(reader *MockGameReader) {
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return([]*game_entities.Game{}, nil)
			},
			validate: func(t *testing.T, games []*game_entities.Game) {
				assert.NotNil(t, games)
				assert.Len(t, games, 0)
			},
		},
		{
			name: "fail when repository returns error",
			setupMocks: func(reader *MockGameReader) {
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(MockGameReader)
			tt.setupMocks(mockReader)

			useCase := usecases.NewSearchGamesUseCase(mockReader)

			ctx := context.Background()
			result, err := useCase.Execute(ctx)

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
