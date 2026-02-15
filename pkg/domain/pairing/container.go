package pairing

import (
	"sync"

	"github.com/golobby/container/v3"
	game_out "github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_in "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/in"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
	schedules_in_ports "github.com/leet-gaming/match-making-api/pkg/domain/schedules/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// mockPoolReader is a simple in-memory implementation for development
type mockPoolReader struct {
	mu    sync.RWMutex
	pools map[string]*pairing_entities.Pool
}

func (m *mockPoolReader) FindPool(criteria *pairing_value_objects.Criteria) (*pairing_entities.Pool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Simple implementation - create a key based on criteria
	key := ""
	if criteria.GameID != nil {
		key += criteria.GameID.String()
	}
	if criteria.Region != nil {
		key += "-" + criteria.Region.Slug
	}
	
	if pool, exists := m.pools[key]; exists {
		return pool, nil
	}
	return nil, nil
}

// mockPoolWriter is a simple in-memory implementation for development
type mockPoolWriter struct {
	reader *mockPoolReader
}

func (m *mockPoolWriter) Save(pool *pairing_entities.Pool) (*pairing_entities.Pool, error) {
	m.reader.mu.Lock()
	defer m.reader.mu.Unlock()
	
	if m.reader.pools == nil {
		m.reader.pools = make(map[string]*pairing_entities.Pool)
	}
	
	// For now, use a simple key - in real implementation this would be based on criteria
	// Since Pool doesn't store criteria, we'll use a default key for development
	key := "default-pool"
	m.reader.pools[key] = pool
	return pool, nil
}

// Inject initializes and registers pairing-related dependencies in the provided container.
//
// Parameters:
//   - container: A container.Container object used as a dependency injection container.
//
// Returns:
//   An error if any initialization or registration fails, otherwise nil.
func Inject(c container.Container) error {
	// Register PartyScheduleMatcher use case
	if err := c.Singleton(func(scheduleReader schedules_in_ports.PartyScheduleReader) (pairing_in.PartyScheduleMatcher, error) {
		return usecases.NewPartyScheduleMatcher(scheduleReader), nil
	}); err != nil {
		return err
	}

	// Register mock PoolReader and PoolWriter for development
	mockReader := &mockPoolReader{}
	if err := c.Singleton(func() pairing_out.PoolReader {
		return mockReader
	}); err != nil {
		return err
	}

	if err := c.Singleton(func() pairing_out.PoolWriter {
		return &mockPoolWriter{reader: mockReader}
	}); err != nil {
		return err
	}

	// Register MatchmakingEventConsumer
	if err := c.Singleton(func(
		addAndFindNextPair *usecases.AddAndFindNextPairUseCase,
		eventPublisher *kafka.EventPublisher,
		regionReader game_out.RegionReader,
		poolReader pairing_out.PoolReader,
		poolWriter pairing_out.PoolWriter,
	) *usecases.MatchmakingEventConsumer {
		return usecases.NewMatchmakingEventConsumer(addAndFindNextPair, eventPublisher, regionReader, poolReader, poolWriter)
	}); err != nil {
		return err
	}

	// Register PlayerQueuedConsumer — consumes PlayerQueued events from matchmaking.commands topic.
	// Wires the MatchmakingEventConsumer.HandlePlayerQueuedProto as the domain handler.
	if err := c.Singleton(func(
		client *kafka.Client,
		eventConsumer *usecases.MatchmakingEventConsumer,
	) *kafka.PlayerQueuedConsumer {
		groupID := "match-making-api-commands"
		return kafka.NewPlayerQueuedConsumer(client, groupID, eventConsumer.HandlePlayerQueuedProto)
	}); err != nil {
		return err
	}

	return nil
}
