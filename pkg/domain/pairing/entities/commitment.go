package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// CommitmentStatus represents the lifecycle state of a player's readiness commitment
type CommitmentStatus int

const (
	CommitmentStatusPending   CommitmentStatus = iota // Awaiting player response
	CommitmentStatusConfirmed                         // Player confirmed readiness
	CommitmentStatusDeclined                          // Player explicitly declined
	CommitmentStatusExpired                           // Commitment window expired without response
	CommitmentStatusTimedOut                          // Timed out by the system worker
)

// String returns a human-readable representation of the commitment status
func (s CommitmentStatus) String() string {
	switch s {
	case CommitmentStatusPending:
		return "pending"
	case CommitmentStatusConfirmed:
		return "confirmed"
	case CommitmentStatusDeclined:
		return "declined"
	case CommitmentStatusExpired:
		return "expired"
	case CommitmentStatusTimedOut:
		return "timed_out"
	default:
		return "unknown"
	}
}

// Commitment tracks an individual player's readiness confirmation for a lobby.
// It serves as the aggregate root for the readiness confirmation bounded context,
// enabling audit trails, channel tracking, and response-time analytics.
type Commitment struct {
	common.BaseEntity
	LobbyID        uuid.UUID              `json:"lobby_id" bson:"lobby_id"`
	PlayerID       uuid.UUID              `json:"player_id" bson:"player_id"`
	Status         CommitmentStatus       `json:"status" bson:"status"`
	Channel        *NotificationChannel   `json:"channel,omitempty" bson:"channel,omitempty"`
	ExpiresAt      time.Time              `json:"expires_at" bson:"expires_at"`
	ConfirmedAt    *time.Time             `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
	DeclinedAt     *time.Time             `json:"declined_at,omitempty" bson:"declined_at,omitempty"`
	ResponseTimeMs *int64                 `json:"response_time_ms,omitempty" bson:"response_time_ms,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// NewCommitment creates a new pending commitment for a player in a lobby.
func NewCommitment(
	resourceOwner common.ResourceOwner,
	lobbyID uuid.UUID,
	playerID uuid.UUID,
	timeoutSeconds int,
	metadata map[string]interface{},
) *Commitment {
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(timeoutSeconds) * time.Second)

	return &Commitment{
		BaseEntity: common.NewEntity(resourceOwner),
		LobbyID:    lobbyID,
		PlayerID:   playerID,
		Status:     CommitmentStatusPending,
		ExpiresAt:  expiresAt,
		Metadata:   metadata,
	}
}

// CanRespond checks whether the player can still respond to this commitment.
func (c *Commitment) CanRespond() bool {
	if c.Status != CommitmentStatusPending {
		return false
	}
	return !c.IsExpired()
}

// IsExpired checks whether the commitment deadline has passed.
func (c *Commitment) IsExpired() bool {
	return time.Now().UTC().After(c.ExpiresAt)
}

// Confirm marks the commitment as confirmed by the player.
func (c *Commitment) Confirm(channel NotificationChannel) error {
	if !c.CanRespond() {
		if c.IsExpired() {
			return fmt.Errorf("commitment has expired")
		}
		return fmt.Errorf("commitment is not pending (status: %s)", c.Status.String())
	}

	now := time.Now().UTC()
	responseMs := now.Sub(c.CreatedAt).Milliseconds()

	c.Status = CommitmentStatusConfirmed
	c.Channel = &channel
	c.ConfirmedAt = &now
	c.ResponseTimeMs = &responseMs
	c.UpdatedAt = now

	return nil
}

// Decline marks the commitment as explicitly declined by the player.
func (c *Commitment) Decline() error {
	if !c.CanRespond() {
		if c.IsExpired() {
			return fmt.Errorf("commitment has expired")
		}
		return fmt.Errorf("commitment is not pending (status: %s)", c.Status.String())
	}

	now := time.Now().UTC()
	responseMs := now.Sub(c.CreatedAt).Milliseconds()

	c.Status = CommitmentStatusDeclined
	c.DeclinedAt = &now
	c.ResponseTimeMs = &responseMs
	c.UpdatedAt = now

	return nil
}

// Expire marks the commitment as expired (called by timeout detection).
func (c *Commitment) Expire() error {
	if c.Status != CommitmentStatusPending {
		return fmt.Errorf("commitment is not pending (status: %s)", c.Status.String())
	}

	now := time.Now().UTC()
	c.Status = CommitmentStatusExpired
	c.UpdatedAt = now

	return nil
}

// TimeOut marks the commitment as timed out by the system worker.
func (c *Commitment) TimeOut() error {
	if c.Status != CommitmentStatusPending {
		return fmt.Errorf("commitment is not pending (status: %s)", c.Status.String())
	}

	now := time.Now().UTC()
	c.Status = CommitmentStatusTimedOut
	c.UpdatedAt = now

	return nil
}

// LobbyCommitmentSummary aggregates all player commitments for a lobby.
// Used for WebSocket broadcasts and frontend countdown rendering.
type LobbyCommitmentSummary struct {
	LobbyID        uuid.UUID                      `json:"lobby_id" bson:"lobby_id"`
	TotalPlayers   int                            `json:"total_players" bson:"total_players"`
	ConfirmedCount int                            `json:"confirmed_count" bson:"confirmed_count"`
	PendingCount   int                            `json:"pending_count" bson:"pending_count"`
	DeclinedCount  int                            `json:"declined_count" bson:"declined_count"`
	ExpiredCount   int                            `json:"expired_count" bson:"expired_count"`
	Deadline       time.Time                      `json:"deadline" bson:"deadline"`
	PlayerStatuses map[uuid.UUID]CommitmentStatus `json:"player_statuses" bson:"player_statuses"`
}

// NewLobbyCommitmentSummary computes a summary from a slice of commitments.
func NewLobbyCommitmentSummary(lobbyID uuid.UUID, commitments []*Commitment) *LobbyCommitmentSummary {
	summary := &LobbyCommitmentSummary{
		LobbyID:        lobbyID,
		TotalPlayers:   len(commitments),
		PlayerStatuses: make(map[uuid.UUID]CommitmentStatus, len(commitments)),
	}

	for _, c := range commitments {
		summary.PlayerStatuses[c.PlayerID] = c.Status
		switch c.Status {
		case CommitmentStatusConfirmed:
			summary.ConfirmedCount++
		case CommitmentStatusPending:
			summary.PendingCount++
		case CommitmentStatusDeclined:
			summary.DeclinedCount++
		case CommitmentStatusExpired, CommitmentStatusTimedOut:
			summary.ExpiredCount++
		}
		if summary.Deadline.IsZero() || c.ExpiresAt.After(summary.Deadline) {
			summary.Deadline = c.ExpiresAt
		}
	}

	return summary
}

// AllConfirmed returns true if every player has confirmed readiness.
func (s *LobbyCommitmentSummary) AllConfirmed() bool {
	return s.TotalPlayers > 0 && s.ConfirmedCount == s.TotalPlayers
}

// HasDeclinedOrExpired returns true if any player declined or timed out.
func (s *LobbyCommitmentSummary) HasDeclinedOrExpired() bool {
	return s.DeclinedCount > 0 || s.ExpiredCount > 0
}
