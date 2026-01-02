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
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_in "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/in"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
	parties_entities "github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
	parties_out "github.com/leet-gaming/match-making-api/pkg/domain/parties/ports/out"
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

// MockPortInvitationWriter is a mock implementation of pairing_out.InvitationWriter using testify/mock
type MockPortInvitationWriter struct {
	mock.Mock
}

// Ensure MockPortInvitationWriter implements pairing_out.InvitationWriter
var _ pairing_out.InvitationWriter = (*MockPortInvitationWriter)(nil)

func (m *MockPortInvitationWriter) Save(ctx context.Context, invitation *pairing_entities.Invitation) (*pairing_entities.Invitation, error) {
	args := m.Called(ctx, invitation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pairing_entities.Invitation), args.Error(1)
}

// MockPortInvitationReader is a mock implementation of pairing_out.InvitationReader using testify/mock
type MockPortInvitationReader struct {
	mock.Mock
}

// Ensure MockPortInvitationReader implements pairing_out.InvitationReader
var _ pairing_out.InvitationReader = (*MockPortInvitationReader)(nil)

func (m *MockPortInvitationReader) GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Invitation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pairing_entities.Invitation), args.Error(1)
}

func (m *MockPortInvitationReader) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*pairing_entities.Invitation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pairing_entities.Invitation), args.Error(1)
}

func (m *MockPortInvitationReader) FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]*pairing_entities.Invitation, error) {
	args := m.Called(ctx, matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pairing_entities.Invitation), args.Error(1)
}

// MockPortPairReader is a mock implementation of pairing_out.PairReader using testify/mock
type MockPortPairReader struct {
	mock.Mock
}

// Ensure MockPortPairReader implements pairing_out.PairReader
var _ pairing_out.PairReader = (*MockPortPairReader)(nil)

func (m *MockPortPairReader) FindPairsByPartyID(ctx context.Context, partyID uuid.UUID) ([]*pairing_entities.Pair, error) {
	args := m.Called(ctx, partyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pairing_entities.Pair), args.Error(1)
}

func (m *MockPortPairReader) GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Pair, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pairing_entities.Pair), args.Error(1)
}

// MockPortPeerReader is a mock implementation of parties_out.PeerReader using testify/mock
type MockPortPeerReader struct {
	mock.Mock
}

// Ensure MockPortPeerReader implements parties_out.PeerReader
var _ parties_out.PeerReader = (*MockPortPeerReader)(nil)

func (m *MockPortPeerReader) GetByID(id uuid.UUID) (*parties_entities.Peer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*parties_entities.Peer), args.Error(1)
}

// MockPoolReader is a mock implementation of pairing_out.PoolReader using testify/mock
type MockPoolReader struct {
	mock.Mock
}

// Ensure MockPoolReader implements pairing_out.PoolReader
var _ pairing_out.PoolReader = (*MockPoolReader)(nil)

func (m *MockPoolReader) FindPool(criteria *pairing_value_objects.Criteria) (*pairing_entities.Pool, error) {
	args := m.Called(criteria)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pairing_entities.Pool), args.Error(1)
}

// MockPoolWriter is a mock implementation of pairing_out.PoolWriter using testify/mock
type MockPoolWriter struct {
	mock.Mock
}

// Ensure MockPoolWriter implements pairing_out.PoolWriter
var _ pairing_out.PoolWriter = (*MockPoolWriter)(nil)

func (m *MockPoolWriter) Save(p *pairing_entities.Pool) (*pairing_entities.Pool, error) {
	args := m.Called(p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pairing_entities.Pool), args.Error(1)
}

// MockPoolInitiator is a mock implementation of pairing_in.PoolInitiator using testify/mock
type MockPoolInitiator struct {
	mock.Mock
}

// Ensure MockPoolInitiator implements pairing_in.PoolInitiator
var _ pairing_in.PoolInitiator = (*MockPoolInitiator)(nil)

func (m *MockPoolInitiator) Execute(c pairing_value_objects.Criteria) (*pairing_entities.Pool, error) {
	args := m.Called(c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pairing_entities.Pool), args.Error(1)
}

// MockPairCreator is a mock implementation of pairing_in.PairCreator using testify/mock
type MockPairCreator struct {
	mock.Mock
}

// Ensure MockPairCreator implements pairing_in.PairCreator
var _ pairing_in.PairCreator = (*MockPairCreator)(nil)

func (m *MockPairCreator) Execute(ctx context.Context, pids []uuid.UUID) (*pairing_entities.Pair, error) {
	args := m.Called(ctx, pids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pairing_entities.Pair), args.Error(1)
}
