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

func TestGetRegionByIDUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		regionID      uuid.UUID
		setupMocks    func(*mocks.MockRegionReader)
		expectedError string
		validate      func(*testing.T, *game_entities.Region)
	}{
		{
			name:     "successfully get region by id",
			regionID: uuid.New(),
			setupMocks: func(reader *mocks.MockRegionReader) {
				region := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Region",
					Slug:       "test-region",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(region, nil)
			},
			validate: func(t *testing.T, region *game_entities.Region) {
				assert.NotNil(t, region)
				assert.Equal(t, "Test Region", region.Name)
				assert.Equal(t, "test-region", region.Slug)
			},
		},
		{
			name:     "fail when region not found",
			regionID: uuid.New(),
			setupMocks: func(reader *mocks.MockRegionReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "failed to get region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(mocks.MockRegionReader)
			tt.setupMocks(mockReader)

			useCase := usecases.NewGetRegionByIDUseCase(mockReader)

			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.regionID)

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
