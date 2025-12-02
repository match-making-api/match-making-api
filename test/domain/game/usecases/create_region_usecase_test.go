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

// MockRegionWriter is a mock implementation of out.RegionWriter
type MockRegionWriter struct {
	mock.Mock
}

func (m *MockRegionWriter) Create(ctx context.Context, region *game_entities.Region) (*game_entities.Region, error) {
	args := m.Called(ctx, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Region), args.Error(1)
}

func (m *MockRegionWriter) Update(ctx context.Context, region *game_entities.Region) (*game_entities.Region, error) {
	args := m.Called(ctx, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Region), args.Error(1)
}

func (m *MockRegionWriter) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockRegionReader is a mock implementation of out.RegionReader
type MockRegionReader struct {
	mock.Mock
}

func (m *MockRegionReader) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.Region, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Region), args.Error(1)
}

func (m *MockRegionReader) Search(ctx context.Context, query interface{}) ([]*game_entities.Region, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_entities.Region), args.Error(1)
}

func TestCreateRegionUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		region        *game_entities.Region
		setupMocks    func(*MockRegionWriter, *MockRegionReader)
		expectedError string
		validate      func(*testing.T, *game_entities.Region)
	}{
		{
			name: "successfully create region",
			region: &game_entities.Region{
				Name:        "Test Region",
				Slug:        "test-region",
				Description: "Test Description",
			},
			setupMocks: func(writer *MockRegionWriter, reader *MockRegionReader) {
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Region{}, nil)
				writer.On("Create", mock.Anything, mock.AnythingOfType("*entities.Region")).Return(func(ctx context.Context, region *game_entities.Region) *game_entities.Region {
					region.ID = uuid.New()
					return region
				}, nil)
			},
			validate: func(t *testing.T, region *game_entities.Region) {
				assert.NotEqual(t, uuid.Nil, region.BaseEntity.ID)
				assert.Equal(t, "Test Region", region.Name)
				assert.Equal(t, "test-region", region.Slug)
			},
		},
		{
			name: "fail when region name is empty",
			region: &game_entities.Region{
				Name: "",
				Slug: "test-region",
			},
			setupMocks: func(writer *MockRegionWriter, reader *MockRegionReader) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "region name is required",
		},
		{
			name: "fail when duplicate region name exists",
			region: &game_entities.Region{
				Name: "Existing Region",
				Slug: "existing-region",
			},
			setupMocks: func(writer *MockRegionWriter, reader *MockRegionReader) {
				existingRegion := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Existing Region",
					Slug:       "existing-region",
				}
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Region{existingRegion}, nil)
			},
			expectedError: "region with name 'Existing Region' already exists",
		},
		{
			name: "fail when duplicate region slug exists",
			region: &game_entities.Region{
				Name: "New Region",
				Slug: "existing-slug",
			},
			setupMocks: func(writer *MockRegionWriter, reader *MockRegionReader) {
				existingRegion := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Existing Region",
					Slug:       "existing-slug",
				}
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Region{existingRegion}, nil)
			},
			expectedError: "region with slug 'existing-slug' already exists",
		},
		{
			name: "fail when repository returns error",
			region: &game_entities.Region{
				Name: "Test Region",
				Slug: "test-region",
			},
			setupMocks: func(writer *MockRegionWriter, reader *MockRegionReader) {
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Region{}, nil)
				writer.On("Create", mock.Anything, mock.AnythingOfType("*entities.Region")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to create region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(MockRegionWriter)
			mockReader := new(MockRegionReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewCreateRegionUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

			result, err := useCase.Execute(ctx, tt.region)

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
