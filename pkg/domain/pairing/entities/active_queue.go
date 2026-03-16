package entities

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ActiveQueueEntry represents a player actively waiting in the matchmaking queue.
// It stores metadata needed for periodic position updates and WebSocket delivery.
type ActiveQueueEntry struct {
	PlayerID        uuid.UUID `json:"player_id"`
	GameID          uuid.UUID `json:"game_id"`
	RegionSlug      string    `json:"region"`
	TenantID        string    `json:"tenant_id"`
	ClientID        string    `json:"client_id"`
	ResourceOwnerID string    `json:"resource_owner_id"`
	Position        int       `json:"position"`
	JoinedAt        time.Time `json:"joined_at"`
}

// InMemoryActiveQueueStore is a thread-safe in-memory implementation of
// pairing_out.ActiveQueueStore. Useful for testing and development.
//
// For production use, prefer RedisActiveQueueStore which supports distributed access.
type InMemoryActiveQueueStore struct {
	mu      sync.RWMutex
	entries map[uuid.UUID]*ActiveQueueEntry
}

// NewInMemoryActiveQueueStore creates a new empty in-memory store.
func NewInMemoryActiveQueueStore() *InMemoryActiveQueueStore {
	return &InMemoryActiveQueueStore{
		entries: make(map[uuid.UUID]*ActiveQueueEntry),
	}
}

func (s *InMemoryActiveQueueStore) Register(_ context.Context, entry *ActiveQueueEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.entries[entry.PlayerID]; ok {
		entry.JoinedAt = existing.JoinedAt
	}

	s.entries[entry.PlayerID] = entry
	return nil
}

func (s *InMemoryActiveQueueStore) Remove(_ context.Context, playerID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, playerID)
	return nil
}

func (s *InMemoryActiveQueueStore) Get(_ context.Context, playerID uuid.UUID) (*ActiveQueueEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries[playerID], nil
}

func (s *InMemoryActiveQueueStore) GetAll(_ context.Context) ([]*ActiveQueueEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ActiveQueueEntry, 0, len(s.entries))
	for _, e := range s.entries {
		cp := *e
		result = append(result, &cp)
	}
	return result, nil
}

func (s *InMemoryActiveQueueStore) Count(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries), nil
}

func (s *InMemoryActiveQueueStore) UpdatePosition(_ context.Context, playerID uuid.UUID, position int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.entries[playerID]; ok {
		entry.Position = position
		return true, nil
	}
	return false, nil
}
