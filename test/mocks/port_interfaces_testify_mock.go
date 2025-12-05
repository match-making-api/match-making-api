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

// MockGameWriter is a mock implementation of out.GameWriter using testify/mock
type MockGameWriter struct {
	mock.Mock
}

// Ensure MockGameWriter implements out.GameWriter
var _ out.GameWriter = (*MockGameWriter)(nil)

func (m *MockGameWriter) Create(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error) {
	args := m.Called(ctx, game)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Game), args.Error(1)
}

func (m *MockGameWriter) Update(ctx context.Context, game *game_entities.Game) (*game_entities.Game, error) {
	args := m.Called(ctx, game)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Game), args.Error(1)
}

func (m *MockGameWriter) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockGameReader is a mock implementation of out.GameReader using testify/mock
type MockGameReader struct {
	mock.Mock
}

// Ensure MockGameReader implements out.GameReader
var _ out.GameReader = (*MockGameReader)(nil)

func (m *MockGameReader) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.Game, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.Game), args.Error(1)
}

func (m *MockGameReader) Search(ctx context.Context, query interface{}) ([]*game_entities.Game, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_entities.Game), args.Error(1)
}

// MockGameModeWriter is a mock implementation of out.GameModeWriter using testify/mock
type MockGameModeWriter struct {
	mock.Mock
}

// Ensure MockGameModeWriter implements out.GameModeWriter
var _ out.GameModeWriter = (*MockGameModeWriter)(nil)

func (m *MockGameModeWriter) Create(ctx context.Context, gameMode *game_entities.GameMode) (*game_entities.GameMode, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.GameMode), args.Error(1)
}

func (m *MockGameModeWriter) Update(ctx context.Context, gameMode *game_entities.GameMode) (*game_entities.GameMode, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.GameMode), args.Error(1)
}

func (m *MockGameModeWriter) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockGameModeReader is a mock implementation of out.GameModeReader using testify/mock
type MockGameModeReader struct {
	mock.Mock
}

// Ensure MockGameModeReader implements out.GameModeReader
var _ out.GameModeReader = (*MockGameModeReader)(nil)

func (m *MockGameModeReader) GetByID(ctx context.Context, id uuid.UUID) (*game_entities.GameMode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game_entities.GameMode), args.Error(1)
}

func (m *MockGameModeReader) Search(ctx context.Context, query interface{}) ([]*game_entities.GameMode, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game_entities.GameMode), args.Error(1)
}

// MockRegionWriter is a mock implementation of out.RegionWriter using testify/mock
type MockRegionWriter struct {
	mock.Mock
}

// Ensure MockRegionWriter implements out.RegionWriter
var _ out.RegionWriter = (*MockRegionWriter)(nil)

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

// MockRegionReader is a mock implementation of out.RegionReader using testify/mock
type MockRegionReader struct {
	mock.Mock
}

// Ensure MockRegionReader implements out.RegionReader
var _ out.RegionReader = (*MockRegionReader)(nil)

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

