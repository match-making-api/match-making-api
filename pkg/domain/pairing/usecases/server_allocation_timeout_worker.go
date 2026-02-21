package usecases

import (
	"context"
	"log/slog"
	"time"

	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// ServerAllocationTimeoutWorkerConfig holds configuration for the timeout worker.
type ServerAllocationTimeoutWorkerConfig struct {
	// TimeoutMinutes is the max wait time before a match is abandoned (default: 10).
	TimeoutMinutes int
	// IntervalSeconds is how often to check for stale entries (default: 60).
	IntervalSeconds int
}

// DefaultServerAllocationTimeoutConfig returns default config.
func DefaultServerAllocationTimeoutConfig() ServerAllocationTimeoutWorkerConfig {
	return ServerAllocationTimeoutWorkerConfig{
		TimeoutMinutes:  10,
		IntervalSeconds: 60,
	}
}

// ServerAllocationTimeoutWorker periodically abandons matches that have been
// waiting for server allocation longer than the configured timepairing_out.
type ServerAllocationTimeoutWorker struct {
	queueStore pairing_out.ServerAllocationQueueStore
	broadcast  WebSocketBroadcastPublisher
	config     ServerAllocationTimeoutWorkerConfig
}

// NewServerAllocationTimeoutWorker creates a new timeout worker.
func NewServerAllocationTimeoutWorker(
	queueStore pairing_out.ServerAllocationQueueStore,
	broadcast WebSocketBroadcastPublisher,
	config ServerAllocationTimeoutWorkerConfig,
) *ServerAllocationTimeoutWorker {
	if config.TimeoutMinutes <= 0 {
		config.TimeoutMinutes = 10
	}
	if config.IntervalSeconds <= 0 {
		config.IntervalSeconds = 60
	}
	return &ServerAllocationTimeoutWorker{
		queueStore: queueStore,
		broadcast:  broadcast,
		config:     config,
	}
}

// Start runs the worker until ctx is cancelled.
func (w *ServerAllocationTimeoutWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(w.config.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.processStale(ctx); err != nil {
				slog.ErrorContext(ctx, "ServerAllocationTimeoutWorker processStale failed", "error", err)
			}
		}
	}
}

func (w *ServerAllocationTimeoutWorker) processStale(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-time.Duration(w.config.TimeoutMinutes) * time.Minute)
	entries, err := w.queueStore.ListStale(ctx, cutoff)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := w.queueStore.Remove(ctx, entry.MatchID); err != nil {
			slog.WarnContext(ctx, "Failed to remove stale match from queue",
				"match_id", entry.MatchID,
				"error", err)
			continue
		}

		payload := &ServerAllocationTimeoutPayload{
			MatchID:         entry.MatchID.String(),
			GameID:         entry.GameID.String(),
			Region:         entry.Region,
			ResourceOwnerID: entry.ResourceOwnerID,
			Reason:         "timeout",
		}

		broadcastEvent := &kafka.WebSocketBroadcastEvent{
			Type:      kafka.EventTypeServerAllocationTimeout,
			LobbyID:   &entry.MatchID,
			TargetIDs: entry.PlayerIDs,
			Payload:   payload,
		}

		if err := w.broadcast.PublishWebSocketBroadcast(ctx, broadcastEvent); err != nil {
			slog.WarnContext(ctx, "Failed to publish ServerAllocationTimeout",
				"match_id", entry.MatchID,
				"error", err)
		} else {
			slog.InfoContext(ctx, "Match abandoned due to server allocation timeout",
				"match_id", entry.MatchID,
				"game_id", entry.GameID,
				"region", entry.Region,
				"wait_minutes", w.config.TimeoutMinutes)
		}
	}

	return nil
}

// ServerAllocationTimeoutPayload is the payload broadcast when a match is abandoned.
type ServerAllocationTimeoutPayload struct {
	MatchID          string `json:"match_id"`
	GameID           string `json:"game_id"`
	Region           string `json:"region"`
	ResourceOwnerID  string `json:"resource_owner_id"`
	Reason           string `json:"reason"`
}
