package usecase

import (
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
)

// LobbyStatusNotice is the match-making payload for a lobby status change.
// replay-api consumes it to refresh UI state. Match-making does not push WebSocket itself.
type LobbyStatusNotice struct {
	LobbyID         uuid.UUID
	Status          string
	Reason          string
	PlayerIDs       []uuid.UUID
	GameType        string
	Region          string
	TenantID        string
	ClientID        string
	ResourceOwnerID string
}

// NewLobbyStatusNotice copies lobby identity, status, seated players, and ownership.
func NewLobbyStatusNotice(lobby *entities.Lobby, reason string) LobbyStatusNotice {
	if lobby == nil {
		return LobbyStatusNotice{Reason: reason}
	}
	lobby.EnsureResourceOwner()
	players := make([]uuid.UUID, 0, len(lobby.PlayerSlots))
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil && !slot.IsSpectator {
			players = append(players, *slot.PlayerID)
		}
	}
	ownerID := ""
	if lobby.ResourceOwner.UserID != uuid.Nil {
		ownerID = lobby.ResourceOwner.UserID.String()
	}
	tenantID := ""
	if lobby.ResourceOwner.TenantID != uuid.Nil {
		tenantID = lobby.ResourceOwner.TenantID.String()
	}
	clientID := ""
	if lobby.ResourceOwner.ClientID != uuid.Nil {
		clientID = lobby.ResourceOwner.ClientID.String()
	}
	return LobbyStatusNotice{
		LobbyID:         lobby.ID,
		Status:          string(lobby.Status),
		Reason:          reason,
		PlayerIDs:       players,
		GameType:        lobby.GameID,
		Region:          lobby.Region,
		TenantID:        tenantID,
		ClientID:        clientID,
		ResourceOwnerID: ownerID,
	}
}
