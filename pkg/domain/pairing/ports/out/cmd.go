package pairing_out

import (
	"context"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
)

type PoolWriter interface {
	Save(p *pairing_entities.Pool) (*pairing_entities.Pool, error)
}

type PairWriter interface {
	Save(p *pairing_entities.Pair) (*pairing_entities.Pair, error)
}

type PairReader interface {
	FindPairsByPartyID(ctx context.Context, partyID uuid.UUID) ([]*pairing_entities.Pair, error)
	GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Pair, error)
}

type InvitationWriter interface {
	Save(ctx context.Context, invitation *pairing_entities.Invitation) (*pairing_entities.Invitation, error)
}

type InvitationReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Invitation, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*pairing_entities.Invitation, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]*pairing_entities.Invitation, error)
}

type PoolReader interface {
	FindPool(criteria *pairing_value_objects.Criteria) (*pairing_entities.Pool, error)
}

type ExternalInvitationWriter interface {
	Save(ctx context.Context, invitation *pairing_entities.ExternalInvitation) (*pairing_entities.ExternalInvitation, error)
}

type ExternalInvitationReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.ExternalInvitation, error)
	GetByRegistrationToken(ctx context.Context, token string) (*pairing_entities.ExternalInvitation, error)
	FindByEmail(ctx context.Context, email string) ([]*pairing_entities.ExternalInvitation, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]*pairing_entities.ExternalInvitation, error)
	FindByEventID(ctx context.Context, eventID uuid.UUID) ([]*pairing_entities.ExternalInvitation, error)
	FindByCreatedBy(ctx context.Context, createdBy uuid.UUID) ([]*pairing_entities.ExternalInvitation, error)
}

type NotificationWriter interface {
	Save(ctx context.Context, notification *pairing_entities.Notification) (*pairing_entities.Notification, error)
	SaveBatch(ctx context.Context, notifications []*pairing_entities.Notification) ([]*pairing_entities.Notification, error)
}

type NotificationReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Notification, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*pairing_entities.Notification, error)
	FindByStatus(ctx context.Context, status pairing_entities.NotificationStatus) ([]*pairing_entities.Notification, error)
	FindFailedNotifications(ctx context.Context) ([]*pairing_entities.Notification, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}

type NotificationTemplateWriter interface {
	Save(ctx context.Context, template *pairing_entities.NotificationTemplate) (*pairing_entities.NotificationTemplate, error)
}

type NotificationTemplateReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.NotificationTemplate, error)
	FindByType(ctx context.Context, notificationType pairing_entities.NotificationType) ([]*pairing_entities.NotificationTemplate, error)
	FindActiveTemplates(ctx context.Context) ([]*pairing_entities.NotificationTemplate, error)
}

type UserNotificationPreferencesWriter interface {
	Save(ctx context.Context, preferences *pairing_entities.UserNotificationPreferences) (*pairing_entities.UserNotificationPreferences, error)
}

type UserNotificationPreferencesReader interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*pairing_entities.UserNotificationPreferences, error)
}

// CommitmentWriter handles persistence of readiness commitments
type CommitmentWriter interface {
	Save(ctx context.Context, commitment *pairing_entities.Commitment) (*pairing_entities.Commitment, error)
	SaveBatch(ctx context.Context, commitments []*pairing_entities.Commitment) ([]*pairing_entities.Commitment, error)
	Update(ctx context.Context, commitment *pairing_entities.Commitment) (*pairing_entities.Commitment, error)
}

// CommitmentReader handles retrieval of readiness commitments
type CommitmentReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Commitment, error)
	FindByLobbyID(ctx context.Context, lobbyID uuid.UUID) ([]*pairing_entities.Commitment, error)
	FindByPlayerAndLobby(ctx context.Context, playerID uuid.UUID, lobbyID uuid.UUID) (*pairing_entities.Commitment, error)
	FindPendingExpired(ctx context.Context) ([]*pairing_entities.Commitment, error)
}

// PushTokenWriter handles persistence of push notification tokens
type PushTokenWriter interface {
	Save(ctx context.Context, token *pairing_entities.PushToken) (*pairing_entities.PushToken, error)
	Deactivate(ctx context.Context, tokenID uuid.UUID) error
}

// PushTokenReader handles retrieval of push notification tokens
type PushTokenReader interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*pairing_entities.PushToken, error)
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*pairing_entities.PushToken, error)
}

// GameConnectionInfoWriter handles persistence of game connection info
type GameConnectionInfoWriter interface {
	Save(ctx context.Context, info *pairing_entities.GameConnectionInfo) (*pairing_entities.GameConnectionInfo, error)
}

// GameConnectionInfoReader handles retrieval of game connection info
type GameConnectionInfoReader interface {
	FindByLobbyID(ctx context.Context, lobbyID uuid.UUID) (*pairing_entities.GameConnectionInfo, error)
	FindByMatchID(ctx context.Context, matchID uuid.UUID) (*pairing_entities.GameConnectionInfo, error)
}

// NotificationChannelSender sends a notification via a specific delivery channel.
type NotificationChannelSender interface {
	Send(ctx context.Context, notification *pairing_entities.Notification) error
	GetChannel() pairing_entities.NotificationChannel
	IsAvailable(ctx context.Context) bool
}

// NotificationSenderResolver resolves the appropriate sender for a notification channel.
// Implemented by NotificationSenderFactory in usecases.
type NotificationSenderResolver interface {
	GetSender(channel pairing_entities.NotificationChannel) NotificationChannelSender
}
