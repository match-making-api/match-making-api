package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// PushPlatform represents the platform a push token belongs to
type PushPlatform string

const (
	PushPlatformWeb     PushPlatform = "web"
	PushPlatformIOS     PushPlatform = "ios"
	PushPlatformAndroid PushPlatform = "android"
)

// PushToken represents a device's push notification registration token (FCM).
type PushToken struct {
	common.BaseEntity
	UserID     uuid.UUID    `json:"user_id" bson:"user_id"`
	Token      string       `json:"token" bson:"token"`
	Platform   PushPlatform `json:"platform" bson:"platform"`
	DeviceName string       `json:"device_name,omitempty" bson:"device_name,omitempty"`
	IsActive   bool         `json:"is_active" bson:"is_active"`
	LastUsedAt *time.Time   `json:"last_used_at,omitempty" bson:"last_used_at,omitempty"`
}

// NewPushToken creates a new active push token entity.
func NewPushToken(
	resourceOwner common.ResourceOwner,
	userID uuid.UUID,
	token string,
	platform PushPlatform,
	deviceName string,
) *PushToken {
	return &PushToken{
		BaseEntity: common.NewEntity(resourceOwner),
		UserID:     userID,
		Token:      token,
		Platform:   platform,
		DeviceName: deviceName,
		IsActive:   true,
	}
}

// Deactivate marks the push token as inactive (e.g., on logout or token rotation).
func (p *PushToken) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now().UTC()
}

// MarkUsed updates the last used timestamp.
func (p *PushToken) MarkUsed() {
	now := time.Now().UTC()
	p.LastUsedAt = &now
	p.UpdatedAt = now
}
