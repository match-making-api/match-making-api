package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
)

type bracketLobbyStore struct {
	lobby *entities.Lobby
}

func (m *bracketLobbyStore) GetByID(ctx context.Context, id uuid.UUID) (*entities.Lobby, error) {
	if m.lobby == nil || m.lobby.ID != id {
		return nil, errors.New("not found")
	}
	return m.lobby, nil
}

func (m *bracketLobbyStore) Update(ctx context.Context, lobby *entities.Lobby) error {
	m.lobby = lobby
	return nil
}

type recordPrizes struct {
	notices []PrizePoolLockedNotice
}

func (r *recordPrizes) PublishPrizePoolLocked(ctx context.Context, notice PrizePoolLockedNotice) error {
	r.notices = append(r.notices, notice)
	return nil
}

type recordHandoff struct {
	calls int
}

func (r *recordHandoff) EnqueueAndBroadcast(ctx context.Context, matchID, gameID uuid.UUID, region string, playerIDs []uuid.UUID, tenantID, clientID, resourceOwnerID string) error {
	r.calls++
	return nil
}

func startedLobby() *entities.Lobby {
	tenant := uuid.New()
	client := uuid.New()
	owner := uuid.New()
	p1 := uuid.New()
	p2 := uuid.New()
	return &entities.Lobby{
		ID:         uuid.New(),
		Type:       entities.LobbyTypeTournament,
		Status:     entities.LobbyStatusOpen,
		GameID:     uuid.NewString(),
		Region:     "br",
		MaxPlayers: 2,
		TenantID:   tenant,
		ClientID:   client,
		CreatorID:  owner,
		ResourceOwner: common.ResourceOwner{
			TenantID: tenant,
			ClientID: client,
			UserID:   owner,
		},
		PrizePool:   &entities.PrizePoolConfig{EntryFeeCents: 100, Status: entities.PrizePoolStatusOpen},
		PlayerSlots: []entities.PlayerSlot{{PlayerID: &p1}, {PlayerID: &p2}},
	}
}

func TestStartTournamentMatch_LocksPoolAndBracket(t *testing.T) {
	lobby := startedLobby()
	store := &bracketLobbyStore{lobby: lobby}
	prizes := &recordPrizes{}
	hand := &recordHandoff{}
	uc := NewStartTournamentMatchUseCase(store, prizes, hand)
	got, err := uc.Execute(context.Background(), StartTournamentMatchInput{
		LobbyID: lobby.ID,
		Caller:  lobby.ResourceOwner,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != entities.LobbyStatusStarting {
		t.Fatalf("status %s", got.Status)
	}
	if got.PrizePool.Status != entities.PrizePoolStatusLocked {
		t.Fatalf("pool %s", got.PrizePool.Status)
	}
	if got.Bracket == nil || got.Bracket.Type != entities.BracketTypeSingleElimination {
		t.Fatal("bracket missing")
	}
	if got.Bracket.Matches[0].Status != entities.BracketMatchStarted || got.Bracket.Matches[0].MatchID == "" {
		t.Fatalf("match %+v", got.Bracket.Matches[0])
	}
	if len(prizes.notices) != 1 || prizes.notices[0].LobbyID != lobby.ID {
		t.Fatalf("prize notices %+v", prizes.notices)
	}
	if hand.calls != 1 {
		t.Fatalf("handoff %d", hand.calls)
	}
}

func TestStartTournamentMatch_SecondStartDoesNotRelock(t *testing.T) {
	lobby := startedLobby()
	lobby.Status = entities.LobbyStatusStarting
	prizes := &recordPrizes{}
	uc := NewStartTournamentMatchUseCase(&bracketLobbyStore{lobby: lobby}, prizes, nil)
	_, err := uc.Execute(context.Background(), StartTournamentMatchInput{
		LobbyID: lobby.ID,
		Caller:  lobby.ResourceOwner,
	})
	if !errors.Is(err, ErrMatchAlreadyStarted) {
		t.Fatalf("got %v", err)
	}
	if len(prizes.notices) != 0 {
		t.Fatal("duplicate prize event")
	}
}

func TestBuildSingleElimination_Bye(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	b := entities.BuildSingleElimination(ids)
	if len(b.Matches) != 2 || b.Matches[1].Status != entities.BracketMatchBye {
		t.Fatalf("%+v", b.Matches)
	}
}
