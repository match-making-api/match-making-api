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

func TestSearchGameModesUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockGameModeReader)
		expectedError string
		validate      func(*testing.T, []*game_entities.GameMode)
	}{
		{
			name: "successfully search game modes",
			setupMocks: func(reader *mocks.MockGameModeReader) {
				gameModes := []*game_entities.GameMode{
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Name:       "Game Mode 1",
					},
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Name:       "Game Mode 2",
					},
				}
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return(gameModes, nil)
			},
			validate: func(t *testing.T, gameModes []*game_entities.GameMode) {
				assert.NotNil(t, gameModes)
				assert.Len(t, gameModes, 2)
			},
		},
		{
			name: "return empty list when no game modes found",
			setupMocks: func(reader *mocks.MockGameModeReader) {
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return([]*game_entities.GameMode{}, nil)
			},
			validate: func(t *testing.T, gameModes []*game_entities.GameMode) {
				assert.NotNil(t, gameModes)
				assert.Len(t, gameModes, 0)
			},
		},
		{
			name: "fail when repository returns error",
			setupMocks: func(reader *mocks.MockGameModeReader) {
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockGameModeReader)
			tt.setupMocks(mockReader)

			useCase := usecases.NewSearchGameModesUseCase(mockReader)

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
