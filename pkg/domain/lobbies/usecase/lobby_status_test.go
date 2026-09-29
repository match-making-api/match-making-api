package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
)

func TestNewLobbyStatusNotice(t *testing.T) {
	tenant := uuid.New()
	client := uuid.New()
	owner := uuid.New()
	player := uuid.New()
	lobby := &entities.Lobby{
		ID:       uuid.New(),
		Status:   entities.LobbyStatusOpen,
		GameID:   "cs2",
		Region:   "br",
		TenantID: tenant,
		ClientID: client,
		ResourceOwner: common.ResourceOwner{
			TenantID: tenant,
			ClientID: client,
			UserID:   owner,
		},
		PlayerSlots: []entities.PlayerSlot{{PlayerID: &player}},
	}
	notice := NewLobbyStatusNotice(lobby, "player_joined")
	if notice.LobbyID != lobby.ID || notice.Status != "open" || notice.Reason != "player_joined" {
		t.Fatalf("%+v", notice)
	}
	if notice.TenantID == "" || notice.ClientID == "" || notice.ResourceOwnerID == "" {
		t.Fatal("ownership required")
	}
	if len(notice.PlayerIDs) != 1 || notice.PlayerIDs[0] != player {
		t.Fatalf("players %+v", notice.PlayerIDs)
	}
}
