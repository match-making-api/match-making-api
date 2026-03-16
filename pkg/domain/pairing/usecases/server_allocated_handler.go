package usecases

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// MatchReadyPayload is the payload broadcast to players when a match is ready.
// replay-api delivers this via WebSocket to each player in TargetIDs.
type MatchReadyPayload struct {
	MatchID         string `json:"match_id"`
	ServerID        string `json:"server_id"`
	Region          string `json:"region"`
	ConnectionURL   string `json:"connection_url,omitempty"`   // Optional; avoid logging sensitive URLs
	ConnectionToken string `json:"connection_token,omitempty"` // Optional; short-lived token
	ResourceOwnerID string `json:"resource_owner_id"`
}

// WebSocketBroadcastPublisher publishes events for WebSocket delivery.
type WebSocketBroadcastPublisher interface {
	PublishWebSocketBroadcast(ctx context.Context, event *kafka.WebSocketBroadcastEvent) error
}

// ServerAllocatedHandler processes ServerAllocated events: validates and broadcasts MatchReady.
type ServerAllocatedHandler struct {
	broadcastPublisher WebSocketBroadcastPublisher
}

// NewServerAllocatedHandler creates a handler for ServerAllocated events.
func NewServerAllocatedHandler(broadcastPublisher WebSocketBroadcastPublisher) *ServerAllocatedHandler {
	return &ServerAllocatedHandler{
		broadcastPublisher: broadcastPublisher,
	}
}

// Handle processes a ServerAllocated event: broadcasts MatchReady to all players in the match.
// Resource ownership is validated by the consumer; only authorized players receive the payload.
func (h *ServerAllocatedHandler) Handle(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.ServerAllocatedPayload) error {
	// Parse player IDs for TargetIDs
	targetIDs := make([]uuid.UUID, 0, len(payload.GetPlayerIds()))
	for _, pidStr := range payload.GetPlayerIds() {
		if pidStr == "" {
			continue
		}
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			slog.WarnContext(ctx, "Invalid player_id in ServerAllocated, skipping",
				"player_id", pidStr,
				"match_id", payload.GetMatchId())
			continue
		}
		targetIDs = append(targetIDs, pid)
	}

	if len(targetIDs) == 0 {
		slog.ErrorContext(ctx, "No valid player_ids for MatchReady broadcast",
			"match_id", payload.GetMatchId(),
			"event_id", envelope.GetId())
		return nil // Don't retry — invalid data
	}

	matchID, err := uuid.Parse(payload.GetMatchId())
	if err != nil {
		slog.ErrorContext(ctx, "Invalid match_id in ServerAllocated",
			"match_id", payload.GetMatchId(),
			"error", err)
		return nil
	}

	matchReadyPayload := &MatchReadyPayload{
		MatchID:         payload.GetMatchId(),
		ServerID:        payload.GetServerId(),
		Region:          payload.GetRegion(),
		ConnectionURL:   payload.GetConnectionUrl(),
		ConnectionToken: payload.GetConnectionToken(),
		ResourceOwnerID: payload.GetResourceOwnerId(),
	}

	broadcastEvent := &kafka.WebSocketBroadcastEvent{
		Type:      kafka.EventTypeMatchReady,
		LobbyID:   &matchID,
		TargetIDs: targetIDs,
		Payload:   matchReadyPayload,
	}

	if err := h.broadcastPublisher.PublishWebSocketBroadcast(ctx, broadcastEvent); err != nil {
		slog.ErrorContext(ctx, "Failed to publish MatchReady broadcast",
			"match_id", payload.GetMatchId(),
			"event_id", envelope.GetId(),
			"error", err)
		return err // Return error to trigger retry
	}

	slog.InfoContext(ctx, "MatchReady broadcast published",
		"match_id", payload.GetMatchId(),
		"server_id", payload.GetServerId(),
		"player_count", len(targetIDs),
		"event_id", envelope.GetId())

	return nil
}
