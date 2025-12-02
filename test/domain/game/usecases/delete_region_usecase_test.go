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

func TestDeleteRegionUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		regionID      uuid.UUID
		setupMocks    func(*MockRegionWriter, *MockRegionReader)
		expectedError string
	}{
		{
			name:     "successfully delete region",
			regionID: uuid.New(),
			setupMocks: func(writer *MockRegionWriter, reader *MockRegionReader) {
				existingRegion := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Region",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingRegion, nil)
				writer.On("Delete", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:     "fail when region not found",
			regionID: uuid.New(),
			setupMocks: func(writer *MockRegionWriter, reader *MockRegionReader) {
				reader.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "region not found",
		},
		{
			name:     "fail when delete fails",
			regionID: uuid.New(),
			setupMocks: func(writer *MockRegionWriter, reader *MockRegionReader) {
				existingRegion := &game_entities.Region{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Test Region",
				}
				reader.On("GetByID", mock.Anything, mock.Anything).Return(existingRegion, nil)
				writer.On("Delete", mock.Anything, mock.Anything).Return(errors.New("delete failed"))
			},
			expectedError: "failed to delete region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(MockRegionWriter)
			mockReader := new(MockRegionReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewDeleteRegionUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

			err := useCase.Execute(ctx, tt.regionID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockWriter.AssertExpectations(t)
			mockReader.AssertExpectations(t)
		})
	}
}
