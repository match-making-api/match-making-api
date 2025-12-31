package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
)

type ExternalInvitationController struct {
	Container container.Container
}

func NewExternalInvitationController(container container.Container) *ExternalInvitationController {
	return &ExternalInvitationController{Container: container}
}

// CreateExternalInvitationRequest represents the request body for creating an external invitation
type CreateExternalInvitationRequest struct {
	Type           string     `json:"type"`             // "match" or "event"
	FullName       string     `json:"full_name"`        // Full name of the external user
	Email          string     `json:"email"`             // Email address of the external user
	Message        string     `json:"message"`           // Invitation message
	ExpirationDate *time.Time `json:"expiration_date,omitempty"` // Expiration date/time
	MatchID        *uuid.UUID `json:"match_id,omitempty"`       // Match ID (if type is "match")
	EventID        *uuid.UUID `json:"event_id,omitempty"`      // Event ID (if type is "event")
}

// Create creates a new external invitation (admin only)
func (eic *ExternalInvitationController) Create(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "method_not_allowed",
				Message: "only POST method is allowed",
			})
			return
		}

		var req CreateExternalInvitationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("invalid JSON: %v", err),
			})
			return
		}

		// Convert type string to ExternalInvitationType
		var invitationType pairing_entities.ExternalInvitationType
		switch req.Type {
		case "match":
			invitationType = pairing_entities.ExternalInvitationTypeMatch
		case "event":
			invitationType = pairing_entities.ExternalInvitationTypeEvent
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "type must be 'match' or 'event'",
			})
			return
		}

		// Resolve dependencies
		var externalInvitationWriter pairing_out.ExternalInvitationWriter
		var externalInvitationReader pairing_out.ExternalInvitationReader
		var pairReader pairing_out.PairReader
		if err := eic.Container.Resolve(&externalInvitationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := eic.Container.Resolve(&externalInvitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := eic.Container.Resolve(&pairReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve PairReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		// Create use case
		createUseCase := &usecases.CreateExternalInvitationUseCase{
			ExternalInvitationWriter: externalInvitationWriter,
			ExternalInvitationReader: externalInvitationReader,
			PairReader:               pairReader,
			Notifier:                 nil, // Can be injected via container if needed
		}

		payload := usecases.CreateExternalInvitationPayload{
			Type:           invitationType,
			FullName:       req.FullName,
			Email:          req.Email,
			Message:        req.Message,
			ExpirationDate: req.ExpirationDate,
			MatchID:        req.MatchID,
			EventID:        req.EventID,
		}

		invitation, err := createUseCase.Execute(r.Context(), payload)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create external invitation", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(invitation)
	}
}

// Get retrieves an external invitation by ID
func (eic *ExternalInvitationController) Get(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		invitationIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "invitation ID is required",
			})
			return
		}

		invitationID, err := uuid.Parse(invitationIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid invitation ID format",
			})
			return
		}

		var externalInvitationReader pairing_out.ExternalInvitationReader
		if err := eic.Container.Resolve(&externalInvitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		getUseCase := &usecases.GetExternalInvitationUseCase{
			ExternalInvitationReader: externalInvitationReader,
		}

		invitation, err := getUseCase.Execute(r.Context(), invitationID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get external invitation", "error", err, "invitation_id", invitationID)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: fmt.Sprintf("invitation not found: %v", err),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitation)
	}
}

// GetByToken retrieves an external invitation by registration token (public endpoint)
func (eic *ExternalInvitationController) GetByToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := r.URL.Query().Get("token")
		if token == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "registration token is required",
			})
			return
		}

		var externalInvitationReader pairing_out.ExternalInvitationReader
		if err := eic.Container.Resolve(&externalInvitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		getByTokenUseCase := &usecases.GetExternalInvitationByTokenUseCase{
			ExternalInvitationReader: externalInvitationReader,
		}

		invitation, err := getByTokenUseCase.Execute(r.Context(), token)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get external invitation by token", "error", err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: fmt.Sprintf("invitation not found: %v", err),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitation)
	}
}

// List lists external invitations with optional filters
func (eic *ExternalInvitationController) List(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var externalInvitationReader pairing_out.ExternalInvitationReader
		if err := eic.Container.Resolve(&externalInvitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		// Get query parameters
		email := r.URL.Query().Get("email")
		matchIDStr := r.URL.Query().Get("match_id")
		eventIDStr := r.URL.Query().Get("event_id")
		createdByStr := r.URL.Query().Get("created_by")
		statusStr := r.URL.Query().Get("status")

		filter := usecases.ListExternalInvitationsFilter{}

		if email != "" {
			filter.Email = &email
		}
		if matchIDStr != "" {
			matchID, err := uuid.Parse(matchIDStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_id",
					Message: "invalid match_id format",
				})
				return
			}
			filter.MatchID = &matchID
		}
		if eventIDStr != "" {
			eventID, err := uuid.Parse(eventIDStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_id",
					Message: "invalid event_id format",
				})
				return
			}
			filter.EventID = &eventID
		}
		if createdByStr != "" {
			createdBy, err := uuid.Parse(createdByStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_id",
					Message: "invalid created_by format",
				})
				return
			}
			filter.CreatedBy = &createdBy
		}
		if statusStr != "" {
			var status pairing_entities.ExternalInvitationStatus
			switch statusStr {
			case "pending":
				status = pairing_entities.ExternalInvitationStatusPending
			case "accepted":
				status = pairing_entities.ExternalInvitationStatusAccepted
			case "expired":
				status = pairing_entities.ExternalInvitationStatusExpired
			case "revoked":
				status = pairing_entities.ExternalInvitationStatusRevoked
			default:
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "validation_error",
					Message: "invalid status value",
				})
				return
			}
			filter.Status = &status
		}

		listUseCase := &usecases.ListExternalInvitationsUseCase{
			ExternalInvitationReader: externalInvitationReader,
		}

		invitations, err := listUseCase.Execute(r.Context(), filter)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to list external invitations", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitations)
	}
}

// Resend resends an external invitation email
func (eic *ExternalInvitationController) Resend(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "method_not_allowed",
				Message: "only POST method is allowed",
			})
			return
		}

		vars := mux.Vars(r)
		invitationIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "invitation ID is required",
			})
			return
		}

		invitationID, err := uuid.Parse(invitationIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid invitation ID format",
			})
			return
		}

		var externalInvitationReader pairing_out.ExternalInvitationReader
		var externalInvitationWriter pairing_out.ExternalInvitationWriter
		if err := eic.Container.Resolve(&externalInvitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := eic.Container.Resolve(&externalInvitationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		resendUseCase := &usecases.ResendExternalInvitationUseCase{
			ExternalInvitationReader: externalInvitationReader,
			ExternalInvitationWriter: externalInvitationWriter,
			Notifier:                 nil, // Can be injected via container if needed
		}

		if err := resendUseCase.Execute(r.Context(), invitationID); err != nil {
			slog.ErrorContext(r.Context(), "failed to resend external invitation", "error", err, "invitation_id", invitationID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Delete revokes an external invitation (admin only)
func (eic *ExternalInvitationController) Delete(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "method_not_allowed",
				Message: "only DELETE method is allowed",
			})
			return
		}

		vars := mux.Vars(r)
		invitationIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "invitation ID is required",
			})
			return
		}

		invitationID, err := uuid.Parse(invitationIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid invitation ID format",
			})
			return
		}

		var externalInvitationReader pairing_out.ExternalInvitationReader
		var externalInvitationWriter pairing_out.ExternalInvitationWriter
		if err := eic.Container.Resolve(&externalInvitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := eic.Container.Resolve(&externalInvitationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve ExternalInvitationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		revokeUseCase := &usecases.RevokeExternalInvitationUseCase{
			ExternalInvitationReader: externalInvitationReader,
			ExternalInvitationWriter: externalInvitationWriter,
			Notifier:                 nil, // Can be injected via container if needed
		}

		if err := revokeUseCase.Execute(r.Context(), invitationID); err != nil {
			slog.ErrorContext(r.Context(), "failed to revoke external invitation", "error", err, "invitation_id", invitationID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
