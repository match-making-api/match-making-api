package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/metrics"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// ReadyCheckTimeoutWorkerConfig holds dependencies for the timeout worker
type ReadyCheckTimeoutWorkerConfig struct {
	CommitmentReader pairing_out.CommitmentReader
	CommitmentWriter pairing_out.CommitmentWriter
	EventPublisher   *EventPublisher
	PollInterval     time.Duration // How often to check for expired commitments (default: 5s)
}

// ReadyCheckTimeoutWorker periodically checks for expired commitments
// and publishes timeout events for them
type ReadyCheckTimeoutWorker struct {
	config *ReadyCheckTimeoutWorkerConfig
}

// NewReadyCheckTimeoutWorker creates a new timeout worker
func NewReadyCheckTimeoutWorker(config *ReadyCheckTimeoutWorkerConfig) *ReadyCheckTimeoutWorker {
	if config.PollInterval == 0 {
		config.PollInterval = 5 * time.Second
	}

	return &ReadyCheckTimeoutWorker{
		config: config,
	}
}

// Start begins the timeout polling loop
func (w *ReadyCheckTimeoutWorker) Start(ctx context.Context) error {
	slog.Info("Starting ready check timeout worker",
		"poll_interval", w.config.PollInterval)

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Ready check timeout worker shutting down")
			return nil
		case <-ticker.C:
			w.processExpiredCommitments(ctx)
		}
	}
}

func (w *ReadyCheckTimeoutWorker) processExpiredCommitments(ctx context.Context) {
	expired, err := w.config.CommitmentReader.FindPendingExpired(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find expired commitments", "error", err)
		return
	}

	if len(expired) == 0 {
		return
	}

	slog.InfoContext(ctx, "Found expired commitments", "count", len(expired))

	// Group by lobby
	lobbyMap := make(map[uuid.UUID][]uuid.UUID)
	for _, commitment := range expired {
		commitment.TimeOut()
		if _, err := w.config.CommitmentWriter.Update(ctx, commitment); err != nil {
			slog.ErrorContext(ctx, "Failed to update timed out commitment",
				"error", err,
				"commitment_id", commitment.ID,
				"lobby_id", commitment.LobbyID)
			continue
		}

		lobbyMap[commitment.LobbyID] = append(lobbyMap[commitment.LobbyID], commitment.PlayerID)
	}

	// Publish timeout events per lobby
	for lobbyID, playerIDs := range lobbyMap {
		metrics.Global.RecordReadyCheckTimedOut(lobbyID.String(), len(playerIDs))

		event := &ReadyCheckEvent{
			EventID:   uuid.New(),
			LobbyID:   lobbyID,
			EventType: EventTypeReadyCheckTimeout,
			PlayerIDs: playerIDs,
			Timestamp: time.Now().UnixMilli(),
			Metadata: map[string]string{
				"timed_out_count": fmt.Sprintf("%d", len(playerIDs)),
			},
		}

		if err := w.config.EventPublisher.PublishReadyCheckEvent(ctx, event); err != nil {
			slog.ErrorContext(ctx, "Failed to publish timeout event",
				"error", err,
				"lobby_id", lobbyID)
		}
	}
}
