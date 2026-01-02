package entities

import (
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// UserNotificationPreferences represents a user's notification preferences
type UserNotificationPreferences struct {
	common.BaseEntity
	UserID              uuid.UUID              `json:"user_id" bson:"user_id"`                           // User ID
	EnabledChannels     []NotificationChannel  `json:"enabled_channels" bson:"enabled_channels"`          // Channels enabled by user
	DisabledChannels    []NotificationChannel  `json:"disabled_channels" bson:"disabled_channels"`       // Channels disabled by user
	DoNotDisturbStart   *string                `json:"do_not_disturb_start,omitempty" bson:"do_not_disturb_start,omitempty"` // DND start time (HH:MM format)
	DoNotDisturbEnd     *string                `json:"do_not_disturb_end,omitempty" bson:"do_not_disturb_end,omitempty"`     // DND end time (HH:MM format)
	PreferredLanguage   string                 `json:"preferred_language" bson:"preferred_language"`     // Preferred language for notifications
	TypePreferences     map[NotificationType]bool `json:"type_preferences" bson:"type_preferences"`        // Preferences per notification type
}

// NewUserNotificationPreferences creates a new user notification preferences entity
func NewUserNotificationPreferences(
	resourceOwner common.ResourceOwner,
	userID uuid.UUID,
	preferredLanguage string,
) *UserNotificationPreferences {
	return &UserNotificationPreferences{
		BaseEntity:          common.NewEntity(resourceOwner),
		UserID:              userID,
		EnabledChannels:     []NotificationChannel{NotificationChannelInApp, NotificationChannelEmail},
		DisabledChannels:    []NotificationChannel{},
		PreferredLanguage:   preferredLanguage,
		TypePreferences:     make(map[NotificationType]bool),
	}
}

// IsChannelEnabled checks if a channel is enabled for the user
func (p *UserNotificationPreferences) IsChannelEnabled(channel NotificationChannel) bool {
	// Check if channel is explicitly disabled
	for _, disabled := range p.DisabledChannels {
		if disabled == channel {
			return false
		}
	}
	// Check if channel is explicitly enabled
	for _, enabled := range p.EnabledChannels {
		if enabled == channel {
			return true
		}
	}
	// Default: in-app is enabled, others are disabled
	return channel == NotificationChannelInApp
}

// IsTypeEnabled checks if a notification type is enabled for the user
func (p *UserNotificationPreferences) IsTypeEnabled(notificationType NotificationType) bool {
	if enabled, exists := p.TypePreferences[notificationType]; exists {
		return enabled
	}
	// Default: all types are enabled
	return true
}
