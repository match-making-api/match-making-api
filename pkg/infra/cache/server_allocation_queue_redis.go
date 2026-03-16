package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// Ensure RedisServerAllocationQueueStore implements the port interface.
var _ pairing_out.ServerAllocationQueueStore = (*RedisServerAllocationQueueStore)(nil)

const (
	serverAllocationQueuePrefix = "matchmaking:server_allocation_queue:"
	serverAllocationIndexPrefix = "matchmaking:server_allocation_index:"
	serverAllocationTTL         = 30 * time.Minute // Max wait before timeout worker cleans
)

// RedisServerAllocationQueueStore is a Redis-backed FIFO queue for matches waiting for server allocation.
// Key structure:
//   - matchmaking:server_allocation_queue:{game_id}:{region} → Redis List (FIFO)
//   - matchmaking:server_allocation_index:{match_id} → hash with game_id, region, enqueued_at (for GetByMatchID, Remove, ListStale)
func queueKey(gameID, region string) string {
	return serverAllocationQueuePrefix + gameID + ":" + region
}

func indexKey(matchID string) string {
	return serverAllocationIndexPrefix + matchID
}

type RedisServerAllocationQueueStore struct {
	client *redis.Client
}

// NewRedisServerAllocationQueueStore creates a new Redis-backed ServerAllocationQueueStore.
func NewRedisServerAllocationQueueStore(client *redis.Client) *RedisServerAllocationQueueStore {
	return &RedisServerAllocationQueueStore{client: client}
}

func (s *RedisServerAllocationQueueStore) Enqueue(ctx context.Context, entry *entities.ServerAllocationQueueEntry) error {
	gk := queueKey(entry.GameID.String(), entry.Region)
	ik := indexKey(entry.MatchID.String())

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.RPush(ctx, gk, entry.MatchID.String())
	pipe.Expire(ctx, gk, serverAllocationTTL)
	pipe.HSet(ctx, ik, "game_id", entry.GameID.String(), "region", entry.Region, "enqueued_at", entry.EnqueuedAt.Unix(), "data", data)
	pipe.Expire(ctx, ik, serverAllocationTTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("enqueue match %s: %w", entry.MatchID, err)
	}
	return nil
}

func (s *RedisServerAllocationQueueStore) DequeueNext(ctx context.Context, gameID uuid.UUID, region string) (*entities.ServerAllocationQueueEntry, error) {
	gk := queueKey(gameID.String(), region)
	matchIDStr, err := s.client.LPop(ctx, gk).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("dequeue: %w", err)
	}

	ik := indexKey(matchIDStr)
	data, err := s.client.HGet(ctx, ik, "data").Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get entry: %w", err)
	}

	var entry entities.ServerAllocationQueueEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal entry: %w", err)
	}

	_ = s.client.Del(ctx, ik).Err()
	return &entry, nil
}

func (s *RedisServerAllocationQueueStore) PeekNext(ctx context.Context, gameID uuid.UUID, region string) (*entities.ServerAllocationQueueEntry, error) {
	gk := queueKey(gameID.String(), region)
	matchIDStr, err := s.client.LIndex(ctx, gk, 0).Result()
	if err == redis.Nil || matchIDStr == "" {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("peek: %w", err)
	}

	ik := indexKey(matchIDStr)
	data, err := s.client.HGet(ctx, ik, "data").Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get entry: %w", err)
	}

	var entry entities.ServerAllocationQueueEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal entry: %w", err)
	}
	return &entry, nil
}

func (s *RedisServerAllocationQueueStore) Remove(ctx context.Context, matchID uuid.UUID) error {
	ik := indexKey(matchID.String())
	gameID, err := s.client.HGet(ctx, ik, "game_id").Result()
	if err == redis.Nil {
		return nil // Already removed
	}
	if err != nil {
		return fmt.Errorf("get index: %w", err)
	}

	region, _ := s.client.HGet(ctx, ik, "region").Result()
	gk := queueKey(gameID, region)
	if err := s.client.LRem(ctx, gk, 0, matchID.String()).Err(); err != nil {
		return fmt.Errorf("remove from queue: %w", err)
	}
	if err := s.client.Del(ctx, ik).Err(); err != nil {
		return fmt.Errorf("remove index: %w", err)
	}
	return nil
}

func (s *RedisServerAllocationQueueStore) GetByMatchID(ctx context.Context, matchID uuid.UUID) (*entities.ServerAllocationQueueEntry, error) {
	ik := indexKey(matchID.String())
	data, err := s.client.HGet(ctx, ik, "data").Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get by match id: %w", err)
	}

	var entry entities.ServerAllocationQueueEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal entry: %w", err)
	}
	return &entry, nil
}

func (s *RedisServerAllocationQueueStore) ListStale(ctx context.Context, cutoff time.Time) ([]*entities.ServerAllocationQueueEntry, error) {
	pattern := serverAllocationIndexPrefix + "*"
	var result []*entities.ServerAllocationQueueEntry
	iter := s.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		ts, err := s.client.HGet(ctx, key, "enqueued_at").Int64()
		if err != nil {
			continue
		}
		if time.Unix(ts, 0).Before(cutoff) {
			data, err := s.client.HGet(ctx, key, "data").Bytes()
			if err != nil {
				continue
			}
			var entry entities.ServerAllocationQueueEntry
			if json.Unmarshal(data, &entry) == nil {
				result = append(result, &entry)
			}
		}
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	return result, nil
}
