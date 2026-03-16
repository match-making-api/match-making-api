package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// CommitmentController handles readiness confirmation / commitment endpoints
type CommitmentController struct {
	Container container.Container
}

// NewCommitmentController creates a new commitment controller
func NewCommitmentController(container container.Container) *CommitmentController {
	return &CommitmentController{Container: container}
}

// ConfirmReadinessRequest represents the optional request body for confirming readiness
type ConfirmReadinessRequest struct {
	Channel string `json:"channel,omitempty"` // Optional: channel used for confirmation
}

// DeclineReadinessRequest represents the optional request body for declining readiness
type DeclineReadinessRequest struct {
	Reason string `json:"reason,omitempty"` // Optional: reason for declining
}

// ConfirmReadiness handles POST /api/lobbies/{lobby_id}/commitments/confirm
func (cc *CommitmentController) ConfirmReadiness(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", Message: "invalid lobby_id"})
			return
		}

		resourceOwner := common.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "authentication required"})
			return
		}

		// Parse optional request body
		var req ConfirmReadinessRequest
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			json.NewDecoder(r.Body).Decode(&req) // Ignore errors — all fields optional
		}

		// Resolve dependencies from container
		var commitmentReader pairing_out.CommitmentReader
		var commitmentWriter pairing_out.CommitmentWriter
		var eventPublisher *kafka.EventPublisher
		var notificationWriter pairing_out.NotificationWriter
		var preferencesReader pairing_out.UserNotificationPreferencesReader
		var senderFactory *usecases.NotificationSenderFactory

		container.Resolve(&commitmentReader)
		container.Resolve(&commitmentWriter)
		container.Resolve(&eventPublisher)
		container.Resolve(&notificationWriter)
		container.Resolve(&preferencesReader)
		container.Resolve(&senderFactory)

		uc := &usecases.ConfirmReadinessUseCase{
			CommitmentReader: commitmentReader,
			CommitmentWriter: commitmentWriter,
			EventPublisher:   eventPublisher,
			NotificationBatch: &usecases.SendBatchNotificationUseCase{
				NotificationWriter:                notificationWriter,
				UserNotificationPreferencesReader: preferencesReader,
				SenderFactory:                     senderFactory,
			},
		}

		result, err := uc.Execute(r.Context(), usecases.ConfirmReadinessPayload{
			LobbyID:  lobbyID,
			PlayerID: resourceOwner.UserID,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to confirm readiness", "error", err, "lobby_id", lobbyID)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "confirmation_failed", Message: err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

// DeclineReadiness handles POST /api/lobbies/{lobby_id}/commitments/decline
func (cc *CommitmentController) DeclineReadiness(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", Message: "invalid lobby_id"})
			return
		}

		resourceOwner := common.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "authentication required"})
			return
		}

		// Parse optional body
		var req DeclineReadinessRequest
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			json.NewDecoder(r.Body).Decode(&req)
		}

		var commitmentReader pairing_out.CommitmentReader
		var commitmentWriter pairing_out.CommitmentWriter
		var eventPublisher *kafka.EventPublisher
		var notificationWriter pairing_out.NotificationWriter
		var preferencesReader pairing_out.UserNotificationPreferencesReader
		var senderFactory *usecases.NotificationSenderFactory

		container.Resolve(&commitmentReader)
		container.Resolve(&commitmentWriter)
		container.Resolve(&eventPublisher)
		container.Resolve(&notificationWriter)
		container.Resolve(&preferencesReader)
		container.Resolve(&senderFactory)

		uc := &usecases.DeclineReadinessUseCase{
			CommitmentReader: commitmentReader,
			CommitmentWriter: commitmentWriter,
			EventPublisher:   eventPublisher,
			NotificationBatch: &usecases.SendBatchNotificationUseCase{
				NotificationWriter:                notificationWriter,
				UserNotificationPreferencesReader: preferencesReader,
				SenderFactory:                     senderFactory,
			},
		}

		summary, err := uc.Execute(r.Context(), usecases.DeclineReadinessPayload{
			LobbyID:  lobbyID,
			PlayerID: resourceOwner.UserID,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to decline readiness", "error", err, "lobby_id", lobbyID)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "decline_failed", Message: err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(summary)
	}
}

// GetCommitmentSummary handles GET /api/lobbies/{lobby_id}/commitments
func (cc *CommitmentController) GetCommitmentSummary(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", Message: "invalid lobby_id"})
			return
		}

		resourceOwner := common.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "authentication required"})
			return
		}

		var commitmentReader pairing_out.CommitmentReader
		container.Resolve(&commitmentReader)

		commitments, err := commitmentReader.FindByLobbyID(r.Context(), lobbyID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get commitments", "error", err, "lobby_id", lobbyID)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "fetch_failed", Message: err.Error()})
			return
		}

		summary := pairing_entities.NewLobbyCommitmentSummary(lobbyID, commitments)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(summary)
	}
}

// GetGameConnectionInfo handles GET /api/lobbies/{lobby_id}/connection-info
func (cc *CommitmentController) GetGameConnectionInfo(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		lobbyID, err := uuid.Parse(mux.Vars(r)["lobby_id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", Message: "invalid lobby_id"})
			return
		}

		resourceOwner := common.GetResourceOwner(r.Context())
		if resourceOwner.UserID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "authentication required"})
			return
		}

		// Verify requesting user is a participant with confirmed status
		var commitmentReader pairing_out.CommitmentReader
		container.Resolve(&commitmentReader)

		commitment, err := commitmentReader.FindByPlayerAndLobby(r.Context(), resourceOwner.UserID, lobbyID)
		if err != nil || commitment == nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "you are not a participant in this lobby"})
			return
		}

		if commitment.Status != pairing_entities.CommitmentStatusConfirmed {
			w.WriteHeader(http.StatusPreconditionFailed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "not_confirmed", Message: "you must confirm readiness first"})
			return
		}

		var connInfoReader pairing_out.GameConnectionInfoReader
		container.Resolve(&connInfoReader)

		connInfo, err := connInfoReader.FindByLobbyID(r.Context(), lobbyID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "not_found", Message: "game connection info not available yet"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(connInfo)
	}
}
