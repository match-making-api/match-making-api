package usecases_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	google_uuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/usecases"
)

func TestGetGameModesUseCase_Execute(t *testing.T) {
	tests := []struct {
		name       string
		gameID     google_uuid.UUID
		setupMocks func(*MockGameModeReader)
		validate   func(*testing.T, []*game_entities.GameMode)
	}{
		{
			name:   "successfully get game modes by game id",
			gameID: google_uuid.New(),
			setupMocks: func(reader *MockGameModeReader) {
				gameModes := []*game_entities.GameMode{
					{
						BaseEntity: common.BaseEntity{ID: google_uuid.New()},
						GameID:     uuid.FromStringOrNil(google_uuid.New().String()),
						Name:       "Game Mode 1",
					},
					{
						BaseEntity: common.BaseEntity{ID: google_uuid.New()},
						GameID:     uuid.FromStringOrNil(google_uuid.New().String()),
						Name:       "Game Mode 2",
					},
				}
				reader.On("Search", mock.Anything, mock.Anything).Return(gameModes, nil)
			},
			validate: func(t *testing.T, gameModes []*game_entities.GameMode) {
				assert.NotNil(t, gameModes)
				assert.Len(t, gameModes, 2)
			},
		},
		{
			name:   "return empty list when no game modes found",
			gameID: google_uuid.New(),
			setupMocks: func(reader *MockGameModeReader) {
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.GameMode{}, nil)
			},
			validate: func(t *testing.T, gameModes []*game_entities.GameMode) {
				assert.NotNil(t, gameModes)
				assert.Len(t, gameModes, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(MockGameModeReader)
			tt.setupMocks(mockReader)

			useCase := usecases.NewGetGameModesUseCase(mockReader)

			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.gameID)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			if tt.validate != nil {
				tt.validate(t, result)
			}

			mockReader.AssertExpectations(t)
		})
	}
}
