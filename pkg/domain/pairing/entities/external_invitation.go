package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// ExternalInvitationStatus represents the current status of an external invitation
type ExternalInvitationStatus int

const (
	ExternalInvitationStatusPending ExternalInvitationStatus = iota
	ExternalInvitationStatusAccepted
	ExternalInvitationStatusExpired
	ExternalInvitationStatusRevoked
)

// ExternalInvitationType represents the type of external invitation
type ExternalInvitationType int

const (
	ExternalInvitationTypeMatch ExternalInvitationType = iota
	ExternalInvitationTypeEvent
)

// ExternalInvitation represents a manual invitation for an external user (not yet on the platform)
// to join a match or event and become a platform member
type ExternalInvitation struct {
	common.BaseEntity
	Type             ExternalInvitationType    `json:"type" bson:"type"`
	FullName         string                    `json:"full_name" bson:"full_name"`                   // Full name of the external user
	Email            string                    `json:"email" bson:"email"`                             // Email address of the external user
	Message          string                    `json:"message" bson:"message"`                        // Invitation message
	ExpirationDate   *time.Time                `json:"expiration_date,omitempty" bson:"expiration_date,omitempty"` // When the invitation expires
	Status           ExternalInvitationStatus   `json:"status" bson:"status"`                          // Current status of the invitation
	RegistrationToken string                   `json:"registration_token" bson:"registration_token"`  // Unique token for registration link
	MatchID          *uuid.UUID                `json:"match_id,omitempty" bson:"match_id,omitempty"`  // Match ID (if type is Match)
	EventID          *uuid.UUID                `json:"event_id,omitempty" bson:"event_id,omitempty"`   // Event ID (if type is Event)
	CreatedBy        uuid.UUID                 `json:"created_by" bson:"created_by"`                   // Administrator who created the invitation
	AcceptedAt       *time.Time                `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
	RevokedAt        *time.Time                `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
	RevokedBy        *uuid.UUID                `json:"revoked_by,omitempty" bson:"revoked_by,omitempty"`
	RegisteredUserID *uuid.UUID                `json:"registered_user_id,omitempty" bson:"registered_user_id,omitempty"` // User ID after registration
}

// NewExternalInvitation creates a new external invitation entity
func NewExternalInvitation(
	resourceOwner common.ResourceOwner,
	invitationType ExternalInvitationType,
	fullName string,
	email string,
	message string,
	expirationDate *time.Time,
	registrationToken string,
	matchID *uuid.UUID,
	eventID *uuid.UUID,
	createdBy uuid.UUID,
) *ExternalInvitation {
	return &ExternalInvitation{
		BaseEntity:        common.NewEntity(resourceOwner),
		Type:              invitationType,
		FullName:          fullName,
		Email:             email,
		Message:          message,
		ExpirationDate:    expirationDate,
		Status:            ExternalInvitationStatusPending,
		RegistrationToken: registrationToken,
		MatchID:           matchID,
		EventID:           eventID,
		CreatedBy:         createdBy,
	}
}

// IsExpired checks if the invitation has expired
func (ei *ExternalInvitation) IsExpired() bool {
	if ei.ExpirationDate == nil {
		return false // No expiration date means it never expires
	}
	return time.Now().After(*ei.ExpirationDate)
}

// CanAccept checks if the invitation can be accepted
func (ei *ExternalInvitation) CanAccept() bool {
	if ei.Status != ExternalInvitationStatusPending {
		return false
	}
	return !ei.IsExpired()
}

// CanRevoke checks if the invitation can be revoked
func (ei *ExternalInvitation) CanRevoke() bool {
	return ei.Status == ExternalInvitationStatusPending
}
