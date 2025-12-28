package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// InvitationStatus represents the current status of an invitation
type InvitationStatus int

const (
	InvitationStatusPending InvitationStatus = iota
	InvitationStatusAccepted
	InvitationStatusDeclined
	InvitationStatusExpired
	InvitationStatusRevoked
)

// InvitationType represents the type of invitation
type InvitationType int

const (
	InvitationTypeMatch InvitationType = iota
	InvitationTypeEvent
)

// Invitation represents a manual invitation for a user to join a match or event
type Invitation struct {
	common.BaseEntity
	Type           InvitationType  `json:"type" bson:"type"`
	UserID         uuid.UUID       `json:"user_id" bson:"user_id"`           // The user being invited
	MatchID        *uuid.UUID      `json:"match_id,omitempty" bson:"match_id,omitempty"` // Match ID (if type is Match)
	EventID        *uuid.UUID      `json:"event_id,omitempty" bson:"event_id,omitempty"` // Event ID (if type is Event)
	Message        string          `json:"message" bson:"message"`           // Invitation message
	ExpirationDate *time.Time      `json:"expiration_date,omitempty" bson:"expiration_date,omitempty"` // When the invitation expires
	Status         InvitationStatus `json:"status" bson:"status"`            // Current status of the invitation
	CreatedBy      uuid.UUID       `json:"created_by" bson:"created_by"`     // Administrator who created the invitation
	AcceptedAt     *time.Time      `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
	DeclinedAt     *time.Time      `json:"declined_at,omitempty" bson:"declined_at,omitempty"`
	RevokedAt      *time.Time      `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
	RevokedBy      *uuid.UUID      `json:"revoked_by,omitempty" bson:"revoked_by,omitempty"`
}

// NewInvitation creates a new invitation entity
func NewInvitation(
	resourceOwner common.ResourceOwner,
	invitationType InvitationType,
	userID uuid.UUID,
	matchID *uuid.UUID,
	eventID *uuid.UUID,
	message string,
	expirationDate *time.Time,
	createdBy uuid.UUID,
) *Invitation {
	return &Invitation{
		BaseEntity:     common.NewEntity(resourceOwner),
		Type:           invitationType,
		UserID:         userID,
		MatchID:        matchID,
		EventID:        eventID,
		Message:        message,
		ExpirationDate: expirationDate,
		Status:         InvitationStatusPending,
		CreatedBy:      createdBy,
	}
}

// IsExpired checks if the invitation has expired
func (i *Invitation) IsExpired() bool {
	if i.ExpirationDate == nil {
		return false // No expiration date means it never expires
	}
	return time.Now().After(*i.ExpirationDate)
}

// CanAccept checks if the invitation can be accepted
func (i *Invitation) CanAccept() bool {
	if i.Status != InvitationStatusPending {
		return false
	}
	return !i.IsExpired()
}

// CanDecline checks if the invitation can be declined
func (i *Invitation) CanDecline() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}

// CanRevoke checks if the invitation can be revoked
func (i *Invitation) CanRevoke() bool {
	return i.Status == InvitationStatusPending
}
