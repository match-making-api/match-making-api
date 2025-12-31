package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
)

type NotificationController struct {
	Container container.Container
}

func NewNotificationController(container container.Container) *NotificationController {
	return &NotificationController{Container: container}
}

// SendNotificationRequest represents the request body for sending a notification
type SendNotificationRequest struct {
	UserID     uuid.UUID              `json:"user_id"`
	Channel    string                 `json:"channel"` // "in_app", "email", "sms"
	Type       string                 `json:"type"`    // "match_invitation", "match_acceptance", "event_reminder", etc.
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	TemplateID *uuid.UUID             `json:"template_id,omitempty"`
	Language   string                 `json:"language,omitempty"`
	MaxRetries int                    `json:"max_retries,omitempty"`
}

// SendBatchNotificationRequest represents the request body for sending batch notifications
type SendBatchNotificationRequest struct {
	UserIDs    []uuid.UUID            `json:"user_ids"`
	Channel    string                 `json:"channel"`
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	TemplateID *uuid.UUID             `json:"template_id,omitempty"`
	Language   string                 `json:"language,omitempty"`
	MaxRetries int                    `json:"max_retries,omitempty"`
}

// Send sends a single notification
func (nc *NotificationController) Send(ctx context.Context) http.HandlerFunc {
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

		var req SendNotificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("invalid JSON: %v", err),
			})
			return
		}

		// Convert channel string to NotificationChannel
		var channel pairing_entities.NotificationChannel
		switch req.Channel {
		case "in_app":
			channel = pairing_entities.NotificationChannelInApp
		case "email":
			channel = pairing_entities.NotificationChannelEmail
		case "sms":
			channel = pairing_entities.NotificationChannelSMS
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "channel must be 'in_app', 'email', or 'sms'",
			})
			return
		}

		// Convert type string to NotificationType
		var notificationType pairing_entities.NotificationType
		switch req.Type {
		case "match_invitation":
			notificationType = pairing_entities.NotificationTypeMatchInvitation
		case "match_acceptance":
			notificationType = pairing_entities.NotificationTypeMatchAcceptance
		case "event_reminder":
			notificationType = pairing_entities.NotificationTypeEventReminder
		case "event_cancellation":
			notificationType = pairing_entities.NotificationTypeEventCancellation
		case "system_announcement":
			notificationType = pairing_entities.NotificationTypeSystemAnnouncement
		case "custom":
			notificationType = pairing_entities.NotificationTypeCustom
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "invalid notification type",
			})
			return
		}

		// Resolve dependencies
		var notificationWriter pairing_out.NotificationWriter
		var notificationReader pairing_out.NotificationReader
		var preferencesReader pairing_out.UserNotificationPreferencesReader
		if err := nc.Container.Resolve(&notificationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve NotificationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := nc.Container.Resolve(&notificationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve NotificationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := nc.Container.Resolve(&preferencesReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve UserNotificationPreferencesReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		// Create sender factory
		senderFactory := usecases.NewNotificationSenderFactory()

		// Create use case
		sendUseCase := &usecases.SendNotificationUseCase{
			NotificationWriter:                notificationWriter,
			NotificationReader:                notificationReader,
			UserNotificationPreferencesReader: preferencesReader,
			SenderFactory:                    senderFactory,
		}

		payload := usecases.SendNotificationPayload{
			UserID:     req.UserID,
			Channel:    channel,
			Type:       notificationType,
			Title:      req.Title,
			Message:    req.Message,
			Metadata:   req.Metadata,
			TemplateID: req.TemplateID,
			Language:   req.Language,
			MaxRetries: req.MaxRetries,
		}

		notification, err := sendUseCase.Execute(r.Context(), payload)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to send notification", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(notification)
	}
}

// SendBatch sends batch notifications
func (nc *NotificationController) SendBatch(ctx context.Context) http.HandlerFunc {
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

		var req SendBatchNotificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("invalid JSON: %v", err),
			})
			return
		}

		// Convert channel and type (similar to Send method)
		var channel pairing_entities.NotificationChannel
		switch req.Channel {
		case "in_app":
			channel = pairing_entities.NotificationChannelInApp
		case "email":
			channel = pairing_entities.NotificationChannelEmail
		case "sms":
			channel = pairing_entities.NotificationChannelSMS
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "channel must be 'in_app', 'email', or 'sms'",
			})
			return
		}

		var notificationType pairing_entities.NotificationType
		switch req.Type {
		case "match_invitation":
			notificationType = pairing_entities.NotificationTypeMatchInvitation
		case "match_acceptance":
			notificationType = pairing_entities.NotificationTypeMatchAcceptance
		case "event_reminder":
			notificationType = pairing_entities.NotificationTypeEventReminder
		case "event_cancellation":
			notificationType = pairing_entities.NotificationTypeEventCancellation
		case "system_announcement":
			notificationType = pairing_entities.NotificationTypeSystemAnnouncement
		case "custom":
			notificationType = pairing_entities.NotificationTypeCustom
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "invalid notification type",
			})
			return
		}

		// Resolve dependencies (similar to Send)
		var notificationWriter pairing_out.NotificationWriter
		var preferencesReader pairing_out.UserNotificationPreferencesReader
		if err := nc.Container.Resolve(&notificationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve NotificationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := nc.Container.Resolve(&preferencesReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve UserNotificationPreferencesReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		senderFactory := usecases.NewNotificationSenderFactory()

		sendBatchUseCase := &usecases.SendBatchNotificationUseCase{
			NotificationWriter:                notificationWriter,
			UserNotificationPreferencesReader: preferencesReader,
			SenderFactory:                    senderFactory,
		}

		payload := usecases.SendBatchNotificationPayload{
			UserIDs:    req.UserIDs,
			Channel:    channel,
			Type:       notificationType,
			Title:      req.Title,
			Message:    req.Message,
			Metadata:   req.Metadata,
			TemplateID: req.TemplateID,
			Language:   req.Language,
			MaxRetries: req.MaxRetries,
		}

		notifications, errors := sendBatchUseCase.Execute(r.Context(), payload)
		
		response := map[string]interface{}{
			"notifications": notifications,
			"success_count": len(notifications),
			"error_count":   len(errors),
		}
		
		if len(errors) > 0 {
			errorMessages := make([]string, len(errors))
			for i, err := range errors {
				errorMessages[i] = err.Error()
			}
			response["errors"] = errorMessages
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// GetUserNotifications retrieves notifications for a user
func (nc *NotificationController) GetUserNotifications(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		userIDStr, ok := vars["user_id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "user_id is required",
			})
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid user_id format",
			})
			return
		}

		// Get limit and offset from query params
		limit := 20
		offset := 0
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		var notificationReader pairing_out.NotificationReader
		if err := nc.Container.Resolve(&notificationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve NotificationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		getUseCase := &usecases.GetUserNotificationsUseCase{
			NotificationReader: notificationReader,
		}

		result, err := getUseCase.Execute(r.Context(), userID, limit, offset)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get user notifications", "error", err, "user_id", userID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

// MarkAsRead marks a notification as read
func (nc *NotificationController) MarkAsRead(ctx context.Context) http.HandlerFunc {
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
		notificationIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "notification ID is required",
			})
			return
		}

		notificationID, err := uuid.Parse(notificationIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid notification ID format",
			})
			return
		}

		var notificationReader pairing_out.NotificationReader
		var notificationWriter pairing_out.NotificationWriter
		if err := nc.Container.Resolve(&notificationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve NotificationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := nc.Container.Resolve(&notificationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve NotificationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		markReadUseCase := &usecases.MarkNotificationReadUseCase{
			NotificationReader: notificationReader,
			NotificationWriter:  notificationWriter,
		}

		if err := markReadUseCase.Execute(r.Context(), notificationID); err != nil {
			slog.ErrorContext(r.Context(), "failed to mark notification as read", "error", err, "notification_id", notificationID)
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

// Retry retries a failed notification
func (nc *NotificationController) Retry(ctx context.Context) http.HandlerFunc {
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
		notificationIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "notification ID is required",
			})
			return
		}

		notificationID, err := uuid.Parse(notificationIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid notification ID format",
			})
			return
		}

		var notificationReader pairing_out.NotificationReader
		var notificationWriter pairing_out.NotificationWriter
		if err := nc.Container.Resolve(&notificationReader); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve NotificationReader", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}
		if err := nc.Container.Resolve(&notificationWriter); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve NotificationWriter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		senderFactory := usecases.NewNotificationSenderFactory()
		retryUseCase := usecases.NewRetryFailedNotificationUseCase(
			notificationReader,
			notificationWriter,
			senderFactory,
			0, // Use default retry delay
		)

		if err := retryUseCase.Execute(r.Context(), notificationID); err != nil {
			slog.ErrorContext(r.Context(), "failed to retry notification", "error", err, "notification_id", notificationID)
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
