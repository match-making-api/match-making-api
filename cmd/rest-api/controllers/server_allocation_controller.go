package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"

	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// ServerAllocationController handles server allocation queue endpoints.
// Used by game server manager to get the next match when a server becomes available.
type ServerAllocationController struct {
	Container container.Container
}

// NewServerAllocationController creates a new ServerAllocationController.
func NewServerAllocationController(c container.Container) *ServerAllocationController {
	return &ServerAllocationController{Container: c}
}

// NextMatchResponse is the response for GET /server-allocation/next.
type NextMatchResponse struct {
	MatchID         string   `json:"match_id"`
	GameID         string   `json:"game_id"`
	Region         string   `json:"region"`
	PlayerIDs      []string `json:"player_ids"`
	TenantID       string   `json:"tenant_id"`
	ClientID       string   `json:"client_id"`
	ResourceOwnerID string  `json:"resource_owner_id"`
	EnqueuedAtMs   int64    `json:"enqueued_at_ms"`
}

// Next returns the next match waiting for server allocation (FIFO per game/region).
// Query params: game_id (required), region (required).
// Returns 200 with match details, or 204 if queue is empty.
func (sc *ServerAllocationController) Next(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameIDStr := r.URL.Query().Get("game_id")
		region := r.URL.Query().Get("region")

		if gameIDStr == "" || region == "" {
			http.Error(w, `{"error":"game_id and region are required"}`, http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		gameID, err := uuid.Parse(gameIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid game_id"}`, http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		var queueStore pairing_out.ServerAllocationQueueStore
		if err := sc.Container.Resolve(&queueStore); err != nil {
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		entry, err := queueStore.DequeueNext(apiContext, gameID, region)
		if err != nil {
			http.Error(w, `{"error":"failed to dequeue"}`, http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		if entry == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		playerIDs := make([]string, len(entry.PlayerIDs))
		for i, pid := range entry.PlayerIDs {
			playerIDs[i] = pid.String()
		}

		resp := NextMatchResponse{
			MatchID:          entry.MatchID.String(),
			GameID:           entry.GameID.String(),
			Region:           entry.Region,
			PlayerIDs:        playerIDs,
			TenantID:         entry.TenantID,
			ClientID:         entry.ClientID,
			ResourceOwnerID:  entry.ResourceOwnerID,
			EnqueuedAtMs:     entry.EnqueuedAt.UnixMilli(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

