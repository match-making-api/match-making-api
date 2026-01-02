package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// NotificationChannel represents the channel through which a notification is sent
type NotificationChannel int

const (
	NotificationChannelInApp NotificationChannel = iota
	NotificationChannelEmail
	NotificationChannelSMS
)

// NotificationStatus represents the current status of a notification
type NotificationStatus int

const (
	NotificationStatusPending NotificationStatus = iota
	NotificationStatusSent
	NotificationStatusFailed
	NotificationStatusRetrying
)

// NotificationType represents the type of notification
type NotificationType int

const (
	NotificationTypeMatchInvitation NotificationType = iota
	NotificationTypeMatchAcceptance
	NotificationTypeEventReminder
	NotificationTypeEventCancellation
	NotificationTypeSystemAnnouncement
	NotificationTypeCustom
)

// Notification represents a notification sent to a user
type Notification struct {
	common.BaseEntity
	UserID         uuid.UUID              `json:"user_id" bson:"user_id"`                   // User who receives the notification
	Channel        NotificationChannel     `json:"channel" bson:"channel"`                  // Channel used to send the notification
	Type           NotificationType        `json:"type" bson:"type"`                         // Type of notification
	Title          string                 `json:"title" bson:"title"`                        // Notification title
	Message        string                 `json:"message" bson:"message"`                    // Notification message body
	Metadata       map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"` // Additional metadata (match ID, event details, timestamps, etc.)
	Status         NotificationStatus      `json:"status" bson:"status"`                     // Current status
	TemplateID     *uuid.UUID             `json:"template_id,omitempty" bson:"template_id,omitempty"` // Template used (if any)
	Language       string                 `json:"language" bson:"language"`                  // Language/locale for localization
	SentAt         *time.Time             `json:"sent_at,omitempty" bson:"sent_at,omitempty"` // When the notification was sent
	FailedAt       *time.Time             `json:"failed_at,omitempty" bson:"failed_at,omitempty"` // When the notification failed (if failed)
	FailureReason  *string                `json:"failure_reason,omitempty" bson:"failure_reason,omitempty"` // Reason for failure
	RetryCount     int                    `json:"retry_count" bson:"retry_count"`            // Number of retry attempts
	MaxRetries     int                    `json:"max_retries" bson:"max_retries"`           // Maximum retry attempts
	NextRetryAt    *time.Time             `json:"next_retry_at,omitempty" bson:"next_retry_at,omitempty"` // When to retry next
	CreatedBy      *uuid.UUID             `json:"created_by,omitempty" bson:"created_by,omitempty"` // Admin who triggered (if manual)
	ReadAt         *time.Time             `json:"read_at,omitempty" bson:"read_at,omitempty"` // When user read the notification (for in-app)
}

// NewNotification creates a new notification entity
func NewNotification(
	resourceOwner common.ResourceOwner,
	userID uuid.UUID,
	channel NotificationChannel,
	notificationType NotificationType,
	title string,
	message string,
	metadata map[string]interface{},
	language string,
	maxRetries int,
	createdBy *uuid.UUID,
) *Notification {
	return &Notification{
		BaseEntity:  common.NewEntity(resourceOwner),
		UserID:      userID,
		Channel:     channel,
		Type:        notificationType,
		Title:       title,
		Message:     message,
		Metadata:    metadata,
		Status:      NotificationStatusPending,
		Language:    language,
		RetryCount:  0,
		MaxRetries:  maxRetries,
		CreatedBy:   createdBy,
	}
}

// CanRetry checks if the notification can be retried
func (n *Notification) CanRetry() bool {
	if n.Status != NotificationStatusFailed {
		return false
	}
	return n.RetryCount < n.MaxRetries
}

// MarkAsSent marks the notification as sent
func (n *Notification) MarkAsSent() {
	now := time.Now()
	n.Status = NotificationStatusSent
	n.SentAt = &now
	n.UpdatedAt = now
}

// MarkAsFailed marks the notification as failed
func (n *Notification) MarkAsFailed(reason string) {
	now := time.Now()
	n.Status = NotificationStatusFailed
	n.FailedAt = &now
	n.FailureReason = &reason
	n.UpdatedAt = now
}

// ScheduleRetry schedules a retry for the notification
func (n *Notification) ScheduleRetry(nextRetryAt time.Time) {
	n.Status = NotificationStatusRetrying
	n.RetryCount++
	n.NextRetryAt = &nextRetryAt
	n.UpdatedAt = time.Now()
}

// MarkAsRead marks the notification as read (for in-app notifications)
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.ReadAt = &now
	n.UpdatedAt = now
}
