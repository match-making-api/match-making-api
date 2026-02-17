package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// Ensure RedisActiveQueueStore implements the port interface.
var _ pairing_out.ActiveQueueStore = (*RedisActiveQueueStore)(nil)

// RedisActiveQueueStore is a Redis/Dragonfly-backed implementation of ActiveQueueStore.
// Each player entry is stored as a JSON hash value keyed by player ID.
// A Redis Set tracks all active player IDs for efficient GetAll/Count operations.
//
// Key structure:
//
//	matchmaking:active_queue:<player_id>  → JSON of ActiveQueueEntry
//	matchmaking:active_queue:players      → Set of player_id strings
//
// TTL: entries expire after 10 minutes to auto-clean stale data.
type RedisActiveQueueStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisActiveQueueStore creates a new Redis-backed ActiveQueueStore.
func NewRedisActiveQueueStore(client *redis.Client) *RedisActiveQueueStore {
	return &RedisActiveQueueStore{
		client: client,
		ttl:    10 * time.Minute,
	}
}

func (s *RedisActiveQueueStore) Register(ctx context.Context, entry *entities.ActiveQueueEntry) error {
	playerID := entry.PlayerID.String()
	key := PlayerKey(playerID)

	// Preserve original JoinedAt if player already registered
	existing, err := s.Get(ctx, entry.PlayerID)
	if err != nil {
		return fmt.Errorf("failed to check existing entry: %w", err)
	}
	if existing != nil {
		entry.JoinedAt = existing.JoinedAt
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.Set(ctx, key, data, s.ttl)
	pipe.SAdd(ctx, ActiveQueueSetKey, playerID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to register player %s: %w", playerID, err)
	}

	return nil
}

func (s *RedisActiveQueueStore) Remove(ctx context.Context, playerID uuid.UUID) error {
	pid := playerID.String()
	key := PlayerKey(pid)

	pipe := s.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, ActiveQueueSetKey, pid)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove player %s: %w", pid, err)
	}

	return nil
}

func (s *RedisActiveQueueStore) Get(ctx context.Context, playerID uuid.UUID) (*entities.ActiveQueueEntry, error) {
	key := PlayerKey(playerID.String())

	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player %s: %w", playerID, err)
	}

	var entry entities.ActiveQueueEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entry for player %s: %w", playerID, err)
	}

	return &entry, nil
}

func (s *RedisActiveQueueStore) GetAll(ctx context.Context) ([]*entities.ActiveQueueEntry, error) {
	memberIDs, err := s.client.SMembers(ctx, ActiveQueueSetKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active player set: %w", err)
	}

	if len(memberIDs) == 0 {
		return nil, nil
	}

	// Build keys for pipeline get
	keys := make([]string, len(memberIDs))
	for i, pid := range memberIDs {
		keys[i] = PlayerKey(pid)
	}

	values, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to mget active queue entries: %w", err)
	}

	entries := make([]*entities.ActiveQueueEntry, 0, len(values))
	staleIDs := make([]string, 0)

	for i, val := range values {
		if val == nil {
			// Entry expired but still in set — mark for cleanup
			staleIDs = append(staleIDs, memberIDs[i])
			continue
		}

		str, ok := val.(string)
		if !ok {
			slog.Warn("Unexpected value type in active queue", "player_id", memberIDs[i])
			continue
		}

		var entry entities.ActiveQueueEntry
		if err := json.Unmarshal([]byte(str), &entry); err != nil {
			slog.Warn("Failed to unmarshal active queue entry", "player_id", memberIDs[i], "error", err)
			staleIDs = append(staleIDs, memberIDs[i])
			continue
		}

		entries = append(entries, &entry)
	}

	// Cleanup stale entries from the set (TTL expired but set entry remained)
	if len(staleIDs) > 0 {
		ifaces := make([]interface{}, len(staleIDs))
		for i, id := range staleIDs {
			ifaces[i] = id
		}
		if err := s.client.SRem(ctx, ActiveQueueSetKey, ifaces...).Err(); err != nil {
			slog.Warn("Failed to cleanup stale active queue set entries", "error", err)
		}
	}

	return entries, nil
}

func (s *RedisActiveQueueStore) Count(ctx context.Context) (int, error) {
	count, err := s.client.SCard(ctx, ActiveQueueSetKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count active queue: %w", err)
	}
	return int(count), nil
}

func (s *RedisActiveQueueStore) UpdatePosition(ctx context.Context, playerID uuid.UUID, position int) (bool, error) {
	entry, err := s.Get(ctx, playerID)
	if err != nil {
		return false, err
	}
	if entry == nil {
		return false, nil
	}

	entry.Position = position

	data, err := json.Marshal(entry)
	if err != nil {
		return false, fmt.Errorf("failed to marshal updated entry: %w", err)
	}

	key := PlayerKey(playerID.String())
	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return false, fmt.Errorf("failed to update position for player %s: %w", playerID, err)
	}

	return true, nil
}
