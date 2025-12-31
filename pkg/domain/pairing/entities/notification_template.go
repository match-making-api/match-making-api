package entities

import (
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// NotificationTemplate represents a reusable template for notifications
type NotificationTemplate struct {
	common.BaseEntity
	Name        string                 `json:"name" bson:"name"`               // Template name
	Type        NotificationType       `json:"type" bson:"type"`               // Notification type this template is for
	Channels    []NotificationChannel  `json:"channels" bson:"channels"`       // Channels this template supports
	Title       string                 `json:"title" bson:"title"`             // Title template (supports variables)
	Message     string                 `json:"message" bson:"message"`         // Message template (supports variables)
	Variables   []string               `json:"variables" bson:"variables"`     // List of variable names used in template
	Languages   []string               `json:"languages" bson:"languages"`    // Supported languages
	IsActive    bool                   `json:"is_active" bson:"is_active"`     // Whether template is active
	CreatedBy   uuid.UUID              `json:"created_by" bson:"created_by"`  // Admin who created the template
}

// NewNotificationTemplate creates a new notification template entity
func NewNotificationTemplate(
	resourceOwner common.ResourceOwner,
	name string,
	notificationType NotificationType,
	channels []NotificationChannel,
	title string,
	message string,
	variables []string,
	languages []string,
	createdBy uuid.UUID,
) *NotificationTemplate {
	return &NotificationTemplate{
		BaseEntity: common.NewEntity(resourceOwner),
		Name:       name,
		Type:        notificationType,
		Channels:    channels,
		Title:       title,
		Message:     message,
		Variables:   variables,
		Languages:   languages,
		IsActive:    true,
		CreatedBy:   createdBy,
	}
}
