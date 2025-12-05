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

func TestSearchRegionsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockRegionReader)
		expectedError string
		validate      func(*testing.T, []*game_entities.Region)
	}{
		{
			name: "successfully search regions",
			setupMocks: func(reader *mocks.MockRegionReader) {
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
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return(regions, nil)
			},
			validate: func(t *testing.T, regions []*game_entities.Region) {
				assert.NotNil(t, regions)
				assert.Len(t, regions, 2)
			},
		},
		{
			name: "return empty list when no regions found",
			setupMocks: func(reader *mocks.MockRegionReader) {
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return([]*game_entities.Region{}, nil)
			},
			validate: func(t *testing.T, regions []*game_entities.Region) {
				assert.NotNil(t, regions)
				assert.Len(t, regions, 0)
			},
		},
		{
			name: "fail when repository returns error",
			setupMocks: func(reader *mocks.MockRegionReader) {
				reader.On("Search", mock.Anything, mock.AnythingOfType("common.Search")).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockRegionReader)
			tt.setupMocks(mockReader)

			useCase := usecases.NewSearchRegionsUseCase(mockReader)

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
