package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/usecases"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestGetRegionsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name       string
		gameID     uuid.UUID
		setupMocks func(*mocks.MockPortRegionReader)
		validate   func(*testing.T, []*game_entities.Region)
	}{
		{
			name:   "successfully get regions",
			gameID: uuid.New(),
			setupMocks: func(reader *mocks.MockPortRegionReader) {
				regions := []*game_entities.Region{
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Name:       "Region 1",
						Slug:       "region-1",
					},
					{
						BaseEntity: common.BaseEntity{ID: uuid.New()},
						Name:       "Region 2",
						Slug:       "region-2",
					},
				}
				reader.On("Search", mock.Anything, mock.Anything).Return(regions, nil)
			},
			validate: func(t *testing.T, regions []*game_entities.Region) {
				assert.NotNil(t, regions)
				assert.Len(t, regions, 2)
			},
		},
		{
			name:   "return empty list when no regions found",
			gameID: uuid.New(),
			setupMocks: func(reader *mocks.MockPortRegionReader) {
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Region{}, nil)
			},
			validate: func(t *testing.T, regions []*game_entities.Region) {
				assert.NotNil(t, regions)
				assert.Len(t, regions, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockPortRegionReader)
			tt.setupMocks(mockReader)

			useCase := usecases.NewGetRegionsUseCase(mockReader)

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
