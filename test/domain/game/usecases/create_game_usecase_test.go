package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/usecases"
)

// MockGameWriter is a mock implementation of out.GameWriter
type MockGameWriter struct {
	mock.Mock
}

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

// MockGameReader is a mock implementation of out.GameReader
type MockGameReader struct {
	mock.Mock
}

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

func TestCreateGameUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		game          *game_entities.Game
		setupMocks    func(*MockGameWriter, *MockGameReader)
		expectedError string
		validate      func(*testing.T, *game_entities.Game)
	}{
		{
			name: "successfully create game",
			game: &game_entities.Game{
				Name:              "Test Game",
				Description:       "Test Description",
				MinPlayersPerTeam: 1,
				MaxPlayersPerTeam: 5,
				NumberOfTeams:     2,
				MaxDuration:       30 * time.Minute,
			},
			setupMocks: func(writer *MockGameWriter, reader *MockGameReader) {
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Game{}, nil)
				writer.On("Create", mock.Anything, mock.AnythingOfType("*entities.Game")).Return(func(ctx context.Context, game *game_entities.Game) *game_entities.Game {
					game.ID = uuid.New()
					return game
				}, nil)
			},
			validate: func(t *testing.T, game *game_entities.Game) {
				assert.NotEqual(t, uuid.Nil, game.BaseEntity.ID)
				assert.True(t, game.Enabled)
				assert.Equal(t, "Test Game", game.Name)
			},
		},
		{
			name: "fail when game name is empty",
			game: &game_entities.Game{
				Name:              "",
				MinPlayersPerTeam: 1,
				MaxPlayersPerTeam: 5,
				NumberOfTeams:     2,
			},
			setupMocks: func(writer *MockGameWriter, reader *MockGameReader) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "game name is required",
		},
		{
			name: "fail when duplicate game name exists",
			game: &game_entities.Game{
				Name:              "Existing Game",
				MinPlayersPerTeam: 1,
				MaxPlayersPerTeam: 5,
				NumberOfTeams:     2,
			},
			setupMocks: func(writer *MockGameWriter, reader *MockGameReader) {
				existingGame := &game_entities.Game{
					BaseEntity: common.BaseEntity{ID: uuid.New()},
					Name:       "Existing Game",
				}
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Game{existingGame}, nil)
			},
			expectedError: "game with name 'Existing Game' already exists",
		},
		{
			name: "fail when min players is greater than max players",
			game: &game_entities.Game{
				Name:              "Test Game",
				MinPlayersPerTeam: 10,
				MaxPlayersPerTeam: 5,
				NumberOfTeams:     2,
			},
			setupMocks: func(writer *MockGameWriter, reader *MockGameReader) {
				// No mocks needed as validation fails before repository calls
			},
			expectedError: "min_players_per_team cannot be greater than max_players_per_team",
		},
		{
			name: "fail when repository returns error",
			game: &game_entities.Game{
				Name:              "Test Game",
				MinPlayersPerTeam: 1,
				MaxPlayersPerTeam: 5,
				NumberOfTeams:     2,
			},
			setupMocks: func(writer *MockGameWriter, reader *MockGameReader) {
				reader.On("Search", mock.Anything, mock.Anything).Return([]*game_entities.Game{}, nil)
				writer.On("Create", mock.Anything, mock.AnythingOfType("*entities.Game")).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to create game",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(MockGameWriter)
			mockReader := new(MockGameReader)
			tt.setupMocks(mockWriter, mockReader)

			useCase := usecases.NewCreateGameUseCase(mockWriter, mockReader)

			ctx := context.Background()
			ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
			ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())

			result, err := useCase.Execute(ctx, tt.game)

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
