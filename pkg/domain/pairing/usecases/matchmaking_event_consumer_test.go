package usecases_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
	parties_entities "github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

// MockAddAndFindNextPairUseCase is a mock for AddAndFindNextPairUseCase
type MockAddAndFindNextPairUseCase struct {
	mock.Mock
}

func (m *MockAddAndFindNextPairUseCase) Execute(payload usecases.FindPairPayload) (*pairing_entities.Pair, *pairing_entities.Pool, int, error) {
	args := m.Called(payload)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*pairing_entities.Pool), args.Int(2), args.Error(3)
	}
	return args.Get(0).(*pairing_entities.Pair), args.Get(1).(*pairing_entities.Pool), args.Int(2), args.Error(3)
}

// MockEventPublisher is a mock for kafka.EventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishMatchCreated(ctx context.Context, event *kafka.MatchEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestMatchmakingEventConsumer_HandleQueueEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("Queue Joined Event - Successful Match", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "us-east-1"
		region := &game_entities.Region{
			Name: "US East",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueJoined,
			PlayerID:  playerID,
			GameType:  gameID.String(),
			Region:    regionSlug,
			MMR:       1500,
		}

		pool := &pairing_entities.Pool{}
		pair := &pairing_entities.Pair{
			Match: map[uuid.UUID]*parties_entities.Party{
				playerID: {ID: playerID},
				uuid.New(): {ID: uuid.New()},
			},
		}
		pair.ID = uuid.New()

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockAddAndFind.On("Execute", mock.MatchedBy(func(payload usecases.FindPairPayload) bool {
			return payload.PartyID == playerID && payload.Criteria.GameID != nil && *payload.Criteria.GameID == gameID && payload.Criteria.Region == region
		})).Return(pair, pool, 1, nil)
		mockEventPublisher.On("PublishMatchCreated", ctx, mock.MatchedBy(func(e *kafka.MatchEvent) bool {
			return e.MatchID == pair.ID && len(e.PlayerIDs) == 2
		})).Return(nil)

		err := consumer.HandleQueueEvent(ctx, event)

		assert.NoError(t, err)
		mockRegionReader.AssertExpectations(t)
		mockAddAndFind.AssertExpectations(t)
		mockEventPublisher.AssertExpectations(t)
	})

	t.Run("Queue Joined Event - Added to Pool", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "eu-west-1"
		region := &game_entities.Region{
			Name: "EU West",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueJoined,
			PlayerID:  playerID,
			GameType:  gameID.String(),
			Region:    regionSlug,
			MMR:       1200,
		}

		pool := &pairing_entities.Pool{}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockAddAndFind.On("Execute", mock.Anything).Return((*pairing_entities.Pair)(nil), pool, 2, nil)

		err := consumer.HandleQueueEvent(ctx, event)

		assert.NoError(t, err)
		mockRegionReader.AssertExpectations(t)
		mockAddAndFind.AssertExpectations(t)
		mockEventPublisher.AssertNotCalled(t, "PublishMatchCreated", mock.Anything, mock.Anything)
	})

	t.Run("Queue Joined Event - Region Not Found", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "invalid-region"

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueJoined,
			PlayerID:  playerID,
			GameType:  gameID.String(),
			Region:    regionSlug,
			MMR:       1000,
		}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{}, nil)

		err := consumer.HandleQueueEvent(ctx, event)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "region not found")
		mockRegionReader.AssertExpectations(t)
		mockAddAndFind.AssertNotCalled(t, "Execute", mock.Anything)
	})

	t.Run("Queue Joined Event - Invalid Game Type UUID", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		regionSlug := "us-west-1"

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueJoined,
			PlayerID:  playerID,
			GameType:  "invalid-uuid",
			Region:    regionSlug,
			MMR:       1000,
		}

		err := consumer.HandleQueueEvent(ctx, event)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
		mockRegionReader.AssertNotCalled(t, "Search", mock.Anything, mock.Anything)
	})

	t.Run("Queue Left Event - Successful Removal", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "us-central-1"
		region := &game_entities.Region{
			Name: "US Central",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueLeft,
			PlayerID:  playerID,
			GameType:  gameID.String(),
			Region:    regionSlug,
			MMR:       1400,
		}

		pool := &pairing_entities.Pool{
			Parties: []uuid.UUID{playerID},
		}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockPoolReader.On("FindPool", mock.MatchedBy(func(c *pairing_value_objects.Criteria) bool {
			return c.GameID != nil && *c.GameID == gameID && c.Region == region
		})).Return(pool, nil)
		mockPoolWriter.On("Save", pool).Return(pool, nil)

		err := consumer.HandleQueueEvent(ctx, event)

		assert.NoError(t, err)
		mockRegionReader.AssertExpectations(t)
		mockPoolReader.AssertExpectations(t)
		mockPoolWriter.AssertExpectations(t)
	})

	t.Run("Queue Left Event - Pool Not Found", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "ap-south-1"
		region := &game_entities.Region{
			Name: "AP South",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueLeft,
			PlayerID:  playerID,
			GameType:  gameID.String(),
			Region:    regionSlug,
			MMR:       1600,
		}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockPoolReader.On("FindPool", mock.Anything).Return((*pairing_entities.Pool)(nil), nil)

		err := consumer.HandleQueueEvent(ctx, event)

		assert.NoError(t, err)
		mockRegionReader.AssertExpectations(t)
		mockPoolReader.AssertExpectations(t)
		mockPoolWriter.AssertNotCalled(t, "Save", mock.Anything)
	})

	t.Run("Unknown Event Type", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		event := &kafka.QueueEvent{
			EventType: "UNKNOWN_EVENT",
			PlayerID:  uuid.New(),
		}

		err := consumer.HandleQueueEvent(ctx, event)

		assert.NoError(t, err)
		// Should just log and return
	})

	t.Run("Queue Left Event - Region Lookup Error", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "us-central-1"

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueLeft,
			PlayerID:  playerID,
			GameType:  gameID.String(),
			Region:    regionSlug,
			MMR:       1400,
		}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return(nil, fmt.Errorf("database error"))

		err := consumer.HandleQueueEvent(ctx, event)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockRegionReader.AssertExpectations(t)
		mockPoolReader.AssertNotCalled(t, "FindPool", mock.Anything)
	})

	t.Run("Queue Left Event - Pool Save Error", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "us-central-1"
		region := &game_entities.Region{
			Name: "US Central",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueLeft,
			PlayerID:  playerID,
			GameType:  gameID.String(),
			Region:    regionSlug,
			MMR:       1400,
		}

		pool := &pairing_entities.Pool{
			Parties: []uuid.UUID{playerID},
		}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockPoolReader.On("FindPool", mock.MatchedBy(func(c *pairing_value_objects.Criteria) bool {
			return c.GameID != nil && *c.GameID == gameID && c.Region == region
		})).Return(pool, nil)
		mockPoolWriter.On("Save", pool).Return(nil, fmt.Errorf("save error"))

		err := consumer.HandleQueueEvent(ctx, event)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "save error")
		mockRegionReader.AssertExpectations(t)
		mockPoolReader.AssertExpectations(t)
		mockPoolWriter.AssertExpectations(t)
	})

	t.Run("Queue Left Event - Pool Remove Error", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "us-central-1"
		region := &game_entities.Region{
			Name: "US Central",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		event := &kafka.QueueEvent{
			EventType: kafka.EventTypeQueueLeft,
			PlayerID:  playerID,
			GameType:  gameID.String(),
			Region:    regionSlug,
			MMR:       1400,
		}

		// Mock pool finding - pool exists but player not in it
		pool := &pairing_entities.Pool{
			Parties: []uuid.UUID{}, // Empty pool - player not in it
		}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockPoolReader.On("FindPool", mock.MatchedBy(func(c *pairing_value_objects.Criteria) bool {
			return c.GameID != nil && *c.GameID == gameID && c.Region == region
		})).Return(pool, nil)

		// Pool remove should fail since player not in pool
		// No save should be called since remove fails

		err := consumer.HandleQueueEvent(ctx, event)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in pool")

		mockRegionReader.AssertExpectations(t)
		mockPoolReader.AssertExpectations(t)
		// PoolWriter should not be called since remove failed
		mockPoolWriter.AssertNotCalled(t, "Save", mock.Anything)

		mockRegionReader.AssertExpectations(t)
		mockPoolReader.AssertExpectations(t)
		mockPoolWriter.AssertExpectations(t)
	})

	t.Run("Player Joined Lobby Event", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		lobbyID := uuid.New()
		playerIDs := []uuid.UUID{uuid.New(), uuid.New()}

		event := &kafka.LobbyEvent{
			EventType: kafka.EventTypePlayerJoined,
			LobbyID:   lobbyID,
			PlayerIDs: playerIDs,
			GameType:  uuid.New().String(),
			Region:    "us-east-1",
			AvgMMR:    1300,
		}

		err := consumer.HandleLobbyEvent(ctx, event)

		assert.NoError(t, err)
		// Currently just logs, no further processing
	})

	t.Run("Unknown Lobby Event Type", func(t *testing.T) {
		// Setup mocks
		mockAddAndFind := &MockAddAndFindNextPairUseCase{}
		mockEventPublisher := &MockEventPublisher{}
		mockRegionReader := &mocks.MockPortRegionReader{}
		mockPoolReader := &mocks.MockPoolReader{}
		mockPoolWriter := &mocks.MockPoolWriter{}

		consumer := usecases.NewMatchmakingEventConsumer(
			mockAddAndFind,
			mockEventPublisher,
			mockRegionReader,
			mockPoolReader,
			mockPoolWriter,
		)

		event := &kafka.LobbyEvent{
			EventType: "UNKNOWN_LOBBY_EVENT",
			LobbyID:   uuid.New(),
		}

		err := consumer.HandleLobbyEvent(ctx, event)

		assert.NoError(t, err)
		// Should just log and return
	})
}

// --- HandlePlayerQueuedProto tests ---

func newTestConsumer() (*usecases.MatchmakingEventConsumer, *MockAddAndFindNextPairUseCase, *MockEventPublisher, *mocks.MockPortRegionReader) {
	mockAddAndFind := &MockAddAndFindNextPairUseCase{}
	mockEventPublisher := &MockEventPublisher{}
	mockRegionReader := &mocks.MockPortRegionReader{}
	mockPoolReader := &mocks.MockPoolReader{}
	mockPoolWriter := &mocks.MockPoolWriter{}

	consumer := usecases.NewMatchmakingEventConsumer(
		mockAddAndFind,
		mockEventPublisher,
		mockRegionReader,
		mockPoolReader,
		mockPoolWriter,
	)

	return consumer, mockAddAndFind, mockEventPublisher, mockRegionReader
}

func TestMatchmakingEventConsumer_HandlePlayerQueuedProto(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Player added to pool", func(t *testing.T) {
		consumer, mockAddAndFind, _, mockRegionReader := newTestConsumer()

		playerID := uuid.New()
		gameID := uuid.New()
		tenantID := uuid.New()
		clientID := uuid.New()
		resourceOwnerID := uuid.New()
		regionSlug := "us-east-1"

		region := &game_entities.Region{
			Name: "US East",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypePlayerQueued,
			Source:          "replay-api",
			Specversion:     schemas.CloudEventsSpecVersion,
			ResourceOwnerId: resourceOwnerID.String(),
		}

		minMMR := int32(1300)
		maxMMR := int32(1700)
		payload := &schemas.PlayerQueuedPayload{
			PlayerId: playerID.String(),
			GameId:   gameID.String(),
			Region:   regionSlug,
			TenantId: tenantID.String(),
			ClientId: clientID.String(),
			SkillRange: &schemas.SkillRange{
				MinMmr: minMMR,
				MaxMmr: maxMMR,
			},
		}

		pool := &pairing_entities.Pool{}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockAddAndFind.On("Execute", mock.MatchedBy(func(p usecases.FindPairPayload) bool {
			return p.PartyID == playerID &&
				p.Criteria.GameID != nil && *p.Criteria.GameID == gameID &&
				p.Criteria.Region == region &&
				p.Criteria.SkillRange != nil &&
				p.Criteria.SkillRange.MinMMR == 1300 &&
				p.Criteria.SkillRange.MaxMMR == 1700
		})).Return((*pairing_entities.Pair)(nil), pool, 1, nil)

		err := consumer.HandlePlayerQueuedProto(ctx, envelope, payload)

		assert.NoError(t, err)
		mockRegionReader.AssertExpectations(t)
		mockAddAndFind.AssertExpectations(t)
	})

	t.Run("Success - Match found and MatchCreated published", func(t *testing.T) {
		consumer, mockAddAndFind, mockEventPublisher, mockRegionReader := newTestConsumer()

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "eu-west-1"

		region := &game_entities.Region{
			Name: "EU West",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypePlayerQueued,
			Source:          "replay-api",
			Specversion:     schemas.CloudEventsSpecVersion,
			ResourceOwnerId: uuid.New().String(),
		}

		payload := &schemas.PlayerQueuedPayload{
			PlayerId: playerID.String(),
			GameId:   gameID.String(),
			Region:   regionSlug,
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}

		pool := &pairing_entities.Pool{}
		pair := &pairing_entities.Pair{
			Match: map[uuid.UUID]*parties_entities.Party{
				playerID:   {ID: playerID},
				uuid.New(): {ID: uuid.New()},
			},
		}
		pair.ID = uuid.New()

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockAddAndFind.On("Execute", mock.Anything).Return(pair, pool, 1, nil)
		mockEventPublisher.On("PublishMatchCreated", ctx, mock.MatchedBy(func(e *kafka.MatchEvent) bool {
			return e.MatchID == pair.ID && len(e.PlayerIDs) == 2
		})).Return(nil)

		err := consumer.HandlePlayerQueuedProto(ctx, envelope, payload)

		assert.NoError(t, err)
		mockRegionReader.AssertExpectations(t)
		mockAddAndFind.AssertExpectations(t)
		mockEventPublisher.AssertExpectations(t)
	})

	t.Run("Error - Missing resource_owner_id", func(t *testing.T) {
		consumer, _, _, _ := newTestConsumer()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypePlayerQueued,
			ResourceOwnerId: "", // empty
		}

		payload := &schemas.PlayerQueuedPayload{
			PlayerId: uuid.New().String(),
			GameId:   uuid.New().String(),
			Region:   "us-east-1",
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}

		err := consumer.HandlePlayerQueuedProto(ctx, envelope, payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource_owner_id")
	})

	t.Run("Error - Invalid player_id UUID", func(t *testing.T) {
		consumer, _, _, _ := newTestConsumer()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypePlayerQueued,
			ResourceOwnerId: uuid.New().String(),
		}

		payload := &schemas.PlayerQueuedPayload{
			PlayerId: "not-a-uuid",
			GameId:   uuid.New().String(),
			Region:   "us-east-1",
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}

		err := consumer.HandlePlayerQueuedProto(ctx, envelope, payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid player_id")
	})

	t.Run("Error - Invalid game_id UUID", func(t *testing.T) {
		consumer, _, _, _ := newTestConsumer()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypePlayerQueued,
			ResourceOwnerId: uuid.New().String(),
		}

		payload := &schemas.PlayerQueuedPayload{
			PlayerId: uuid.New().String(),
			GameId:   "not-a-uuid",
			Region:   "us-east-1",
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}

		err := consumer.HandlePlayerQueuedProto(ctx, envelope, payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid game_id")
	})

	t.Run("Error - Region not found", func(t *testing.T) {
		consumer, _, _, mockRegionReader := newTestConsumer()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypePlayerQueued,
			ResourceOwnerId: uuid.New().String(),
		}

		payload := &schemas.PlayerQueuedPayload{
			PlayerId: uuid.New().String(),
			GameId:   uuid.New().String(),
			Region:   "nonexistent-region",
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": "nonexistent-region"}).Return([]*game_entities.Region{}, nil)

		err := consumer.HandlePlayerQueuedProto(ctx, envelope, payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "region not found")
		mockRegionReader.AssertExpectations(t)
	})

	t.Run("Error - AddAndFindNextPair fails", func(t *testing.T) {
		consumer, mockAddAndFind, _, mockRegionReader := newTestConsumer()

		playerID := uuid.New()
		gameID := uuid.New()
		regionSlug := "us-east-1"

		region := &game_entities.Region{
			Name: "US East",
			Slug: regionSlug,
		}
		region.ID = uuid.New()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypePlayerQueued,
			ResourceOwnerId: uuid.New().String(),
		}

		payload := &schemas.PlayerQueuedPayload{
			PlayerId: playerID.String(),
			GameId:   gameID.String(),
			Region:   regionSlug,
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}

		mockRegionReader.On("Search", ctx, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
		mockAddAndFind.On("Execute", mock.Anything).Return((*pairing_entities.Pair)(nil), (*pairing_entities.Pool)(nil), -1, fmt.Errorf("pool creation failed"))

		err := consumer.HandlePlayerQueuedProto(ctx, envelope, payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool creation failed")
		mockRegionReader.AssertExpectations(t)
		mockAddAndFind.AssertExpectations(t)
	})
}