package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

func TestBuildTournamentEntryFeeEvent_SkipsFreeEntry(t *testing.T) {
	ev, ok := BuildTournamentEntryFeeEvent(TournamentEntryFeeInput{
		TournamentID: "t1",
		PlayerID:     "p1",
		AmountCents:  0,
	})
	if ok || ev != nil {
		t.Fatalf("free entry must not publish, ok=%v ev=%v", ok, ev)
	}
}

func TestBuildTournamentEntryFeeEvent_Payload(t *testing.T) {
	ev, ok := BuildTournamentEntryFeeEvent(TournamentEntryFeeInput{
		TournamentID:    "tour-1",
		PlayerID:        "player-1",
		AmountCents:     500,
		ResourceOwnerID: "owner-1",
		TenantID:        "tenant-1",
		ClientID:        "client-1",
		CorrelationID:   "corr-1",
		GameID:          "cs2",
		Region:          "br",
	})
	if !ok || ev == nil || ev.ChargeableOperationRequested == nil {
		t.Fatal("expected event")
	}
	p := ev.ChargeableOperationRequested
	if p.OperationType != schemas.OperationTypeTournamentEntry {
		t.Fatalf("operation_type %s", p.OperationType)
	}
	if p.AmountCents != 500 || p.PlayerID != "player-1" || p.TournamentID != "tour-1" {
		t.Fatalf("payload %+v", p)
	}
	if p.ResourceOwnerID != "owner-1" || p.CorrelationID != "corr-1" {
		t.Fatalf("owner/correlation %+v", p)
	}
	if p.IdempotencyKey != "tournament_entry:player-1:tour-1" {
		t.Fatalf("idempotency %s", p.IdempotencyKey)
	}
}

func TestEntryFeeInputFromLobby_OnlyTournamentWithFee(t *testing.T) {
	owner := uuid.New()
	lobbyID := uuid.New()
	player := uuid.New()
	custom := &entities.Lobby{
		ID:        lobbyID,
		Type:      entities.LobbyTypeCustom,
		PrizePool: &entities.PrizePoolConfig{EntryFeeCents: 100},
	}
	if in := EntryFeeInputFromLobby(custom, player, "c"); in.AmountCents != 0 {
		t.Fatalf("custom lobby fee %d", in.AmountCents)
	}
	tournament := &entities.Lobby{
		ID:        lobbyID,
		Type:      entities.LobbyTypeTournament,
		CreatorID: owner,
		GameID:    "cs2",
		Region:    "br",
		PrizePool: &entities.PrizePoolConfig{EntryFeeCents: 250},
	}
	in := EntryFeeInputFromLobby(tournament, player, "corr")
	if in.AmountCents != 250 || in.TournamentID != lobbyID.String() || in.ResourceOwnerID != owner.String() {
		t.Fatalf("input %+v", in)
	}
}
