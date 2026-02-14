package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	parties_out "github.com/leet-gaming/match-making-api/pkg/domain/parties/ports/out"
)

type InvitationController struct {
	Container container.Container
}

func NewInvitationController(container container.Container) *InvitationController {
	return &InvitationController{Container: container}
}

// CreateInvitationRequest represents the request body for creating an invitation
type CreateInvitationRequest struct {
	Type           string     `json:"type"`             // "match" or "event"
	UserID         uuid.UUID  `json:"user_id"`          // User being invited
	MatchID        *uuid.UUID `json:"match_id,omitempty"` // Match ID (if type is "match")
	EventID        *uuid.UUID `json:"event_id,omitempty"` // Event ID (if type is "event")
	Message        string     `json:"message"`          // Invitation message
	ExpirationDate *time.Time `json:"expiration_date,omitempty"` // Expiration date/time
}

// UpdateInvitationRequest represents the request body for updating an invitation
type UpdateInvitationRequest struct {
	Message        *string    `json:"message,omitempty"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// Create creates a new manual invitation (admin only)
func (ic *InvitationController) Create(ctx context.Context) http.HandlerFunc {
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

		// SECURITY: Admin-only endpoint
		if !common.IsAdmin(r.Context()) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "forbidden",
				Message: "administrator access required",
			})
			return
		}

		var req CreateInvitationRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: "invalid request body",
			})
			return
		}

		// Convert type string to InvitationType
		var invitationType pairing_entities.InvitationType
		switch req.Type {
		case "match":
			invitationType = pairing_entities.InvitationTypeMatch
		case "event":
			invitationType = pairing_entities.InvitationTypeEvent
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "type must be 'match' or 'event'",
			})
			return
		}

		// Resolve dependencies
		var invitationWriter pairing_out.InvitationWriter
		var peerReader parties_out.PeerReader
		var pairReader pairing_out.PairReader
		if err := ic.Container.Resolve(&invitationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := ic.Container.Resolve(&peerReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve PeerReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := ic.Container.Resolve(&pairReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve PairReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		// Create use case
		createUseCase := &usecases.CreateManualInvitationUseCase{
			InvitationWriter: invitationWriter,
			PeerReader:       peerReader,
			PairReader:       pairReader,
			Notifier:         nil, // Can be injected via container if needed
		}

		payload := usecases.CreateInvitationPayload{
			Type:           invitationType,
			UserID:         req.UserID,
			MatchID:        req.MatchID,
			EventID:        req.EventID,
			Message:        req.Message,
			ExpirationDate: req.ExpirationDate,
		}

		invitation, err := createUseCase.Execute(r.Context(), payload)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create invitation", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to create invitation",
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(invitation)
	}
}

// Get retrieves an invitation by ID
func (ic *InvitationController) Get(ctx context.Context) http.HandlerFunc {
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

		var invitationReader pairing_out.InvitationReader
		if err := ic.Container.Resolve(&invitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		invitation, err := invitationReader.GetByID(r.Context(), invitationID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get invitation", "error", err, "invitation_id", invitationID)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: "invitation not found",
			})
			return
		}

		// Verify user is owner or admin (security check)
		userIDValue := r.Context().Value(common.UserIDKey)
		currentUserID, ok := userIDValue.(uuid.UUID)
		if !ok || currentUserID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "user ID not found in context",
			})
			return
		}
		if invitation.UserID != currentUserID && !common.IsAdmin(r.Context()) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "forbidden",
				Message: "you do not have permission to access this invitation",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitation)
	}
}

// List lists invitations with optional filters
func (ic *InvitationController) List(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var invitationReader pairing_out.InvitationReader
		if err := ic.Container.Resolve(&invitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		// Get query parameters
		userIDStr := r.URL.Query().Get("user_id")
		matchIDStr := r.URL.Query().Get("match_id")

		var invitations []*pairing_entities.Invitation
		var err error

		if userIDStr != "" {
			userID, parseErr := uuid.Parse(userIDStr)
			if parseErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_id",
					Message: "invalid user_id format",
				})
				return
			}
			// Verify user can access this user_id's invitations (security check)
			userIDValue := r.Context().Value(common.UserIDKey)
			currentUserID, ok := userIDValue.(uuid.UUID)
			if !ok || currentUserID == uuid.Nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "unauthorized",
					Message: "user ID not found in context",
				})
				return
			}
			if userID != currentUserID && !common.IsAdmin(r.Context()) {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "forbidden",
					Message: "you do not have permission to access invitations for this user",
				})
				return
			}
			invitations, err = invitationReader.FindByUserID(r.Context(), userID)
		} else if matchIDStr != "" {
			matchID, parseErr := uuid.Parse(matchIDStr)
			if parseErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "invalid_id",
					Message: "invalid match_id format",
				})
				return
			}
			// Verify admin access (security check)
			if !common.IsAdmin(r.Context()) {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "forbidden",
					Message: "only administrators can access invitations by match_id",
				})
				return
			}
			invitations, err = invitationReader.FindByMatchID(r.Context(), matchID)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "user_id or match_id query parameter is required",
			})
			return
		}

		if err != nil {
			slog.ErrorContext(r.Context(), "failed to list invitations", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to retrieve invitations",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitations)
	}
}

// Accept accepts an invitation
func (ic *InvitationController) Accept(ctx context.Context) http.HandlerFunc {
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

		// Get userID from context (set by authentication middleware)
		userIDValue := r.Context().Value(common.UserIDKey)
		userID, ok := userIDValue.(uuid.UUID)
		if !ok || userID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "user ID not found in context",
			})
			return
		}

		var invitationReader pairing_out.InvitationReader
		var invitationWriter pairing_out.InvitationWriter
		if err := ic.Container.Resolve(&invitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := ic.Container.Resolve(&invitationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		acceptUseCase := &usecases.AcceptInvitationUseCase{
			InvitationReader: invitationReader,
			InvitationWriter: invitationWriter,
			Notifier:         nil, // Can be injected via container if needed
		}

		if err := acceptUseCase.Execute(r.Context(), invitationID, userID); err != nil {
			slog.ErrorContext(r.Context(), "failed to accept invitation", "error", err, "invitation_id", invitationID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to accept invitation",
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Decline declines an invitation
func (ic *InvitationController) Decline(ctx context.Context) http.HandlerFunc {
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

		// Get userID from context (set by authentication middleware)
		userIDValue := r.Context().Value(common.UserIDKey)
		userID, ok := userIDValue.(uuid.UUID)
		if !ok || userID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "user ID not found in context",
			})
			return
		}

		var invitationReader pairing_out.InvitationReader
		var invitationWriter pairing_out.InvitationWriter
		if err := ic.Container.Resolve(&invitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := ic.Container.Resolve(&invitationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		declineUseCase := &usecases.DeclineInvitationUseCase{
			InvitationReader: invitationReader,
			InvitationWriter: invitationWriter,
			Notifier:         nil, // Can be injected via container if needed
		}

		if err := declineUseCase.Execute(r.Context(), invitationID, userID); err != nil {
			slog.ErrorContext(r.Context(), "failed to decline invitation", "error", err, "invitation_id", invitationID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to decline invitation",
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Update updates an invitation (admin only)
func (ic *InvitationController) Update(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Admin-only endpoint
		if !common.IsAdmin(r.Context()) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "forbidden",
				Message: "administrator access required",
			})
			return
		}

		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "method_not_allowed",
				Message: "only PATCH method is allowed",
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

		var req UpdateInvitationRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: "invalid request body",
			})
			return
		}

		var invitationReader pairing_out.InvitationReader
		var invitationWriter pairing_out.InvitationWriter
		if err := ic.Container.Resolve(&invitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := ic.Container.Resolve(&invitationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		updateUseCase := &usecases.UpdateInvitationUseCase{
			InvitationReader: invitationReader,
			InvitationWriter: invitationWriter,
		}

		payload := usecases.UpdateInvitationPayload{
			Message:        req.Message,
			ExpirationDate: req.ExpirationDate,
		}

		invitation, err := updateUseCase.Execute(r.Context(), invitationID, payload)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to update invitation", "error", err, "invitation_id", invitationID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to update invitation",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitation)
	}
}

// Delete revokes an invitation (admin only)
func (ic *InvitationController) Delete(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Admin-only endpoint
		if !common.IsAdmin(r.Context()) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "forbidden",
				Message: "administrator access required",
			})
			return
		}

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

		var invitationReader pairing_out.InvitationReader
		var invitationWriter pairing_out.InvitationWriter
		if err := ic.Container.Resolve(&invitationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := ic.Container.Resolve(&invitationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve InvitationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		revokeUseCase := &usecases.RevokeInvitationUseCase{
			InvitationReader: invitationReader,
			InvitationWriter: invitationWriter,
			Notifier:         nil, // Can be injected via container if needed
		}

		if err := revokeUseCase.Execute(r.Context(), invitationID); err != nil {
			slog.ErrorContext(r.Context(), "failed to revoke invitation", "error", err, "invitation_id", invitationID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to revoke invitation",
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
