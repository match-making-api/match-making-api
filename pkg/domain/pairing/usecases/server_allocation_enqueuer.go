package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// ServerAllocationEnqueuerImpl enqueues matches for server allocation and broadcasts WaitingForServer.
type ServerAllocationEnqueuerImpl struct {
	queueStore pairing_out.ServerAllocationQueueStore
	broadcast  WebSocketBroadcastPublisher
}

// NewServerAllocationEnqueuer creates a new ServerAllocationEnqueuer.
func NewServerAllocationEnqueuer(queueStore pairing_out.ServerAllocationQueueStore, broadcast WebSocketBroadcastPublisher) *ServerAllocationEnqueuerImpl {
	return &ServerAllocationEnqueuerImpl{
		queueStore: queueStore,
		broadcast:  broadcast,
	}
}

// EnqueueAndBroadcast enqueues the match and broadcasts WaitingForServer to all players.
func (e *ServerAllocationEnqueuerImpl) EnqueueAndBroadcast(ctx context.Context, matchID, gameID uuid.UUID, region string, playerIDs []uuid.UUID, tenantID, clientID, resourceOwnerID string) error {
	entry := &entities.ServerAllocationQueueEntry{
		MatchID:         matchID,
		GameID:          gameID,
		Region:          region,
		PlayerIDs:       playerIDs,
		TenantID:        tenantID,
		ClientID:        clientID,
		ResourceOwnerID: resourceOwnerID,
		EnqueuedAt:      time.Now().UTC(),
	}

	if err := e.queueStore.Enqueue(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "Failed to enqueue match for server allocation",
			"match_id", matchID,
			"game_id", gameID,
			"region", region,
			"error", err)
		return err
	}

	payload := &WaitingForServerPayload{
		MatchID:         matchID.String(),
		GameID:          gameID.String(),
		Region:          region,
		ResourceOwnerID: resourceOwnerID,
		EnqueuedAtMs:    entry.EnqueuedAt.UnixMilli(),
	}

	targetIDs := make([]uuid.UUID, len(playerIDs))
	copy(targetIDs, playerIDs)

	broadcastEvent := &kafka.WebSocketBroadcastEvent{
		Type:      kafka.EventTypeWaitingForServer,
		LobbyID:   &matchID,
		TargetIDs: targetIDs,
		Payload:   payload,
	}

	if err := e.broadcast.PublishWebSocketBroadcast(ctx, broadcastEvent); err != nil {
		slog.ErrorContext(ctx, "Failed to publish WaitingForServer broadcast",
			"match_id", matchID,
			"error", err)
		return err
	}

	slog.InfoContext(ctx, "Match enqueued for server allocation",
		"match_id", matchID,
		"game_id", gameID,
		"region", region,
		"player_count", len(playerIDs))

	return nil
}

// WaitingForServerPayload is the payload broadcast when a match is waiting for server allocation.
type WaitingForServerPayload struct {
	MatchID         string `json:"match_id"`
	GameID          string `json:"game_id"`
	Region          string `json:"region"`
	ResourceOwnerID string `json:"resource_owner_id"`
	EnqueuedAtMs    int64  `json:"enqueued_at_ms"`
}
