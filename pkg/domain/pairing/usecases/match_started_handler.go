package usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// MatchStartedBroadcastPayload is the payload broadcast to players when a match is about to begin.
// replay-api delivers this via WebSocket to each player in TargetIDs.
type MatchStartedBroadcastPayload struct {
	MatchID               string `json:"match_id"`
	ResourceOwnerID       string `json:"resource_owner_id"`
	CountdownSeconds      int32  `json:"countdown_seconds,omitempty"`
	StartTimestampEpochMs int64  `json:"start_timestamp_epoch_ms,omitempty"`
}

// MatchStartedHandler processes MatchStarted events: validates and broadcasts to participants.
type MatchStartedHandler struct {
	broadcastPublisher WebSocketBroadcastPublisher
}

// NewMatchStartedHandler creates a handler for MatchStarted events.
func NewMatchStartedHandler(broadcastPublisher WebSocketBroadcastPublisher) *MatchStartedHandler {
	return &MatchStartedHandler{
		broadcastPublisher: broadcastPublisher,
	}
}

// Handle processes a MatchStarted event: broadcasts to all match participants via WebSocket.
// Resource ownership is validated by the consumer; only authorized players receive the payload.
func (h *MatchStartedHandler) Handle(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.MatchStartedPayload) error {
	targetIDs := make([]uuid.UUID, 0, len(payload.GetPlayerIds()))
	for _, pidStr := range payload.GetPlayerIds() {
		if pidStr == "" {
			continue
		}
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			slog.WarnContext(ctx, "Invalid player_id in MatchStarted, skipping",
				"player_id", pidStr,
				"match_id", payload.GetMatchId())
			continue
		}
		targetIDs = append(targetIDs, pid)
	}

	if len(targetIDs) == 0 {
		slog.ErrorContext(ctx, "No valid player_ids for MatchStarted broadcast",
			"match_id", payload.GetMatchId(),
			"event_id", envelope.GetId())
		return nil
	}

	matchID, err := uuid.Parse(payload.GetMatchId())
	if err != nil {
		slog.ErrorContext(ctx, "Invalid match_id in MatchStarted",
			"match_id", payload.GetMatchId(),
			"error", err)
		return nil
	}

	wsPayload := &MatchStartedBroadcastPayload{
		MatchID:               payload.GetMatchId(),
		ResourceOwnerID:       payload.GetResourceOwnerId(),
		CountdownSeconds:      payload.GetCountdownSeconds(),
		StartTimestampEpochMs: payload.GetStartTimestampEpochMs(),
	}

	broadcastEvent := &kafka.WebSocketBroadcastEvent{
		Type:      kafka.EventTypeMatchStarted,
		LobbyID:   &matchID,
		TargetIDs: targetIDs,
		Payload:   wsPayload,
	}

	if err := h.broadcastPublisher.PublishWebSocketBroadcast(ctx, broadcastEvent); err != nil {
		slog.ErrorContext(ctx, "Failed to publish MatchStarted broadcast",
			"match_id", payload.GetMatchId(),
			"event_id", envelope.GetId(),
			"error", err)
		return err
	}

	slog.InfoContext(ctx, "MatchStarted broadcast published",
		"match_id", payload.GetMatchId(),
		"player_count", len(targetIDs),
		"event_id", envelope.GetId())

	return nil
}
