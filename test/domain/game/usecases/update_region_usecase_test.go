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

func TestUpdateRegionUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		regionID      uuid.UUID
		region        *game_entities.Region
		setupMocks    func(*mocks.MockRegionWriter, *mocks.MockRegionReader)
		expectedError string
		validate      func(*testing.T, *game_entities.Region)
	}{
		{
			name:     "successfully update region",
			regionID: uuid.New(),
			region: &game_entities.Region{
				Name:        "Updated Region",
				Slug:        "updated-region",
				Description: "Updated Description",
			},
			setupMocks: func(writer *mocks.MockRegionWriter, reader *mocks.MockRegionReader) {
				existingRegion := &game_entities.Region{
					BaseEntity:  common.BaseEntity{ID: uuid.New()},
					Name:        "Original Region",
					Slug:        "original-region",
					Description: "Original Description",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingRegion, nil)
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Region{}, nil)
				writer.On("Update", mock.Anything, mock.AnythingOfType("*entities.Region")).Return(existingRegion, nil)
			},
			validate: func(t *testing.T, region *game_entities.Region) {
				assert.Equal(t, "Updated Region", region.Name)
				assert.Equal(t, "updated-region", region.Slug)
				assert.Equal(t, "Updated Description", region.Description)
			},
		},
		{
			name:     "fail when region not found",
			regionID: uuid.New(),
			region: &game_entities.Region{
				Name: "Test Region",
				Slug: "test-region",
			},
			setupMocks: func(writer *mocks.MockRegionWriter, reader *mocks.MockRegionReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "region not found",
		},
		{
			name:     "fail when duplicate name exists",
			regionID: uuid.New(),
			region: &game_entities.Region{
				Name: "Duplicate Name",
				Slug: "test-region",
			},
			setupMocks: func(writer *mocks.MockRegionWriter, reader *mocks.MockRegionReader) {
				existingRegion := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Original Region",
					Slug:       "original-region",
				}
				duplicateRegion := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Duplicate Name",
					Slug:       "duplicate-region",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingRegion, nil)
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Region{existingRegion, duplicateRegion}, nil)
			},
			expectedError: "region with name 'Duplicate Name' already exists",
		},
		{
			name:     "fail when duplicate slug exists",
			regionID: uuid.New(),
			region: &game_entities.Region{
				Name: "Test Region",
				Slug: "existing-slug",
			},
			setupMocks: func(writer *mocks.MockRegionWriter, reader *mocks.MockRegionReader) {
				existingRegion := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Original Region",
					Slug:       "original-region",
				}
				duplicateRegion := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Another Region",
					Slug:       "existing-slug",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingRegion, nil)
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Region{existingRegion, duplicateRegion}, nil)
			},
			expectedError: "region with slug 'existing-slug' already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(mocks.MockRegionWriter)
			mockReader := new(mocks.MockRegionReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewUpdateRegionUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

			result, err := useCase.Execute(ctx, tt.regionID, tt.region)

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
