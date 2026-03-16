package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
	game_out "github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// QueueStatusTickerConfig holds configuration for the periodic queue status updater.
type QueueStatusTickerConfig struct {
	// Interval between position update broadcasts (default: 5 seconds)
	Interval time.Duration

	// EstimatedWaitPerPosition is the estimated wait time per position in the queue.
	// Used to derive estimated_wait_ms for the client (default: 10 seconds per position).
	EstimatedWaitPerPosition time.Duration

	// Enabled controls whether the ticker is active. Can be toggled at runtime via feature flag.
	Enabled bool
}

// DefaultQueueStatusTickerConfig returns sensible defaults for the queue status ticker.
func DefaultQueueStatusTickerConfig() *QueueStatusTickerConfig {
	return &QueueStatusTickerConfig{
		Interval:                 5 * time.Second,
		EstimatedWaitPerPosition: 10 * time.Second,
		Enabled:                  true,
	}
}

// QueueStatusTicker periodically scans active queue entries and publishes
// QueueStatusUpdated events via Kafka for WebSocket broadcast to players.
//
// Architecture: match-making-api is the source of truth for pool state, so it
// produces these events. replay-api (or another service) consumes from
// websocket.broadcasts and delivers to the player's WebSocket connection.
type QueueStatusTicker struct {
	activeQueueStore pairing_out.ActiveQueueStore
	poolReader       pairing_out.PoolReader
	regionReader     game_out.RegionReader
	eventPublisher   *kafka.EventPublisher
	config           *QueueStatusTickerConfig
}

// NewQueueStatusTicker creates a new ticker for periodic queue status broadcasts.
func NewQueueStatusTicker(
	activeQueueStore pairing_out.ActiveQueueStore,
	poolReader pairing_out.PoolReader,
	regionReader game_out.RegionReader,
	eventPublisher *kafka.EventPublisher,
	config *QueueStatusTickerConfig,
) *QueueStatusTicker {
	if config == nil {
		config = DefaultQueueStatusTickerConfig()
	}
	return &QueueStatusTicker{
		activeQueueStore: activeQueueStore,
		poolReader:       poolReader,
		regionReader:     regionReader,
		eventPublisher:   eventPublisher,
		config:           config,
	}
}

// Start begins the periodic ticker. It blocks until the context is cancelled.
func (t *QueueStatusTicker) Start(ctx context.Context) error {
	if !t.config.Enabled {
		slog.Info("QueueStatusTicker is disabled, not starting")
		return nil
	}

	slog.Info("Starting QueueStatusTicker",
		"interval", t.config.Interval,
		"estimated_wait_per_position", t.config.EstimatedWaitPerPosition)

	ticker := time.NewTicker(t.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("QueueStatusTicker stopped (context cancelled)")
			return nil
		case <-ticker.C:
			t.tick(ctx)
		}
	}
}

// tick performs a single iteration: scans all active queue entries,
// refreshes positions from pool state, and publishes status events.
func (t *QueueStatusTicker) tick(ctx context.Context) {
	entries, err := t.activeQueueStore.GetAll(ctx)
	if err != nil {
		slog.WarnContext(ctx, "Failed to get active queue entries", "error", err)
		return
	}
	if len(entries) == 0 {
		return
	}

	slog.Debug("QueueStatusTicker tick", "active_players", len(entries))

	for _, entry := range entries {
		// Refresh position from pool state
		position := t.refreshPosition(ctx, entry)
		if position < 0 {
			// Player no longer in pool — remove from active queue
			if err := t.activeQueueStore.Remove(ctx, entry.PlayerID); err != nil {
				slog.WarnContext(ctx, "Failed to remove player from active queue",
					"player_id", entry.PlayerID, "error", err)
			}
			slog.InfoContext(ctx, "Player no longer in pool, removed from active queue",
				"player_id", entry.PlayerID)
			continue
		}

		// Update stored position
		if _, err := t.activeQueueStore.UpdatePosition(ctx, entry.PlayerID, position); err != nil {
			slog.WarnContext(ctx, "Failed to update player position",
				"player_id", entry.PlayerID, "error", err)
		}

		// Compute estimated wait
		estimatedWaitMs := int64(position) * t.config.EstimatedWaitPerPosition.Milliseconds()

		// Count total players in queue for this game/region
		totalInQueue := t.countPlayersInQueue(entries, entry.GameID, entry.RegionSlug)

		// Publish WebSocket broadcast event targeted at this player
		if t.eventPublisher == nil {
			continue
		}

		payload := &kafka.QueueStatusPayload{
			PlayerID:        entry.PlayerID,
			GameID:          entry.GameID,
			Region:          entry.RegionSlug,
			Position:        position,
			EstimatedWaitMs: estimatedWaitMs,
			TotalInQueue:    totalInQueue,
			EventType:       kafka.EventTypeQueueStatusUpdated,
			Timestamp:       time.Now().UnixMilli(),
		}

		broadcastEvent := &kafka.WebSocketBroadcastEvent{
			Type:      kafka.EventTypeQueueStatusUpdated,
			TargetIDs: []uuid.UUID{entry.PlayerID},
			Payload:   payload,
			Timestamp: time.Now().UnixMilli(),
		}

		if err := t.eventPublisher.PublishWebSocketBroadcast(ctx, broadcastEvent); err != nil {
			slog.WarnContext(ctx, "Failed to publish queue status update",
				"player_id", entry.PlayerID,
				"error", err)
		}
	}
}

// refreshPosition looks up the player's current position in the pool.
// Returns -1 if the player is no longer in any pool.
func (t *QueueStatusTicker) refreshPosition(ctx context.Context, entry *pairing_entities.ActiveQueueEntry) int {
	// Build criteria to find the pool
	regions, err := t.regionReader.Search(ctx, map[string]interface{}{"slug": entry.RegionSlug})
	if err != nil || len(regions) == 0 {
		return -1
	}

	criteria := &pairing_value_objects.Criteria{
		GameID: &entry.GameID,
		Region: regions[0],
	}

	pool, err := t.poolReader.FindPool(criteria)
	if err != nil || pool == nil {
		return -1
	}

	pos, found := pool.IsQueued(entry.PlayerID)
	if !found {
		return -1
	}

	return pos + 1 // IsQueued returns 0-based index; position is 1-based
}

// countPlayersInQueue counts how many active entries share the same game and region.
func (t *QueueStatusTicker) countPlayersInQueue(entries []*pairing_entities.ActiveQueueEntry, gameID uuid.UUID, regionSlug string) int {
	count := 0
	for _, e := range entries {
		if e.GameID == gameID && e.RegionSlug == regionSlug {
			count++
		}
	}
	return count
}
