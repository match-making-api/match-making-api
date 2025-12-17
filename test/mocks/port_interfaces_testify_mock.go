// Code generated manually for use with testify/mock.
// This file should not be regenerated automatically.
// To regenerate, update the interfaces in pkg/domain/game/ports/out/cmd.go
// and manually update this file accordingly.

package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

// MockPortGameWriter is a mock implementation of out.GameWriter using testify/mock
// This is for port interfaces, not MongoDB repository implementations
type MockPortGameWriter struct {
	mock.Mock
}

// Ensure MockPortGameWriter implements out.GameWriter
var _ out.GameWriter = (*MockPortGameWriter)(nil)

func (m *MockPortGameWriter) Create(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error) {
	args := m.Called(ctx, game)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Game), args.Error(1)
}

func (m *MockPortGameWriter) Update(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error) {
	args := m.Called(ctx, game)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Game), args.Error(1)
}

func (m *MockPortGameWriter) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPortGameReader is a mock implementation of out.GameReader using testify/mock
// This is for port interfaces, not MongoDB repository implementations
type MockPortGameReader struct {
	mock.Mock
}

// Ensure MockPortGameReader implements out.GameReader
var _ out.GameReader = (*MockPortGameReader)(nil)

func (m *MockPortGameReader) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.Game, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Game), args.Error(1)
}

func (m *MockPortGameReader) Search(ctx context.Context, query interface{}) ([]*game_entities.Game, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_entities.Game), args.Error(1)
}

// MockPortGameModeWriter is a mock implementation of out.GameModeWriter using testify/mock
// This is for port interfaces, not MongoDB repository implementations
type MockPortGameModeWriter struct {
	mock.Mock
}

// Ensure MockPortGameModeWriter implements out.GameModeWriter
var _ out.GameModeWriter = (*MockPortGameModeWriter)(nil)

func (m *MockPortGameModeWriter) Create(ctx context.Context, gameMode *game_entities.GameMode) (*game_entities.GameMode, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.GameMode), args.Error(1)
}

func (m *MockPortGameModeWriter) Update(ctx context.Context, gameMode *game_entities.GameMode) (*game_entities.GameMode, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.GameMode), args.Error(1)
}

func (m *MockPortGameModeWriter) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPortGameModeReader is a mock implementation of out.GameModeReader using testify/mock
// This is for port interfaces, not MongoDB repository implementations
type MockPortGameModeReader struct {
	mock.Mock
}

// Ensure MockPortGameModeReader implements out.GameModeReader
var _ out.GameModeReader = (*MockPortGameModeReader)(nil)

func (m *MockPortGameModeReader) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.GameMode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.GameMode), args.Error(1)
}

func (m *MockPortGameModeReader) Search(ctx context.Context, query interface{}) ([]*game_entities.GameMode, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_entities.GameMode), args.Error(1)
}

// MockPortRegionWriter is a mock implementation of out.RegionWriter using testify/mock
// This is for port interfaces, not MongoDB repository implementations
type MockPortRegionWriter struct {
	mock.Mock
}

// Ensure MockPortRegionWriter implements out.RegionWriter
var _ out.RegionWriter = (*MockPortRegionWriter)(nil)

func (m *MockPortRegionWriter) Create(ctx context.Context, region *game_entities.Region) (*game_entities.Region, error) {
	args := m.Called(ctx, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Region), args.Error(1)
}

func (m *MockPortRegionWriter) Update(ctx context.Context, region *game_entities.Region) (*game_entities.Region, error) {
	args := m.Called(ctx, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Region), args.Error(1)
}

func (m *MockPortRegionWriter) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPortRegionReader is a mock implementation of out.RegionReader using testify/mock
// This is for port interfaces, not MongoDB repository implementations
type MockPortRegionReader struct {
	mock.Mock
}

// Ensure MockPortRegionReader implements out.RegionReader
var _ out.RegionReader = (*MockPortRegionReader)(nil)

func (m *MockPortRegionReader) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.Region, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Region), args.Error(1)
}

func (m *MockPortRegionReader) Search(ctx context.Context, query interface{}) ([]*game_entities.Region, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_entities.Region), args.Error(1)
}
