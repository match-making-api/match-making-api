package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

type memLobbyStore struct {
	lobby *entities.Lobby
}

func (m *memLobbyStore) GetByID(ctx context.Context, id uuid.UUID) (*entities.Lobby, error) {
	if m.lobby == nil || m.lobby.ID != id {
		return nil, errors.New("not found")
	}
	return m.lobby, nil
}

func (m *memLobbyStore) Update(ctx context.Context, lobby *entities.Lobby) error {
	m.lobby = lobby
	return nil
}

type recordJoinPublisher struct {
	fees   []*schemas.ChargeableOperationEvent
	joined []PlayerJoinedNotice
	feeErr error
}

func (r *recordJoinPublisher) PublishChargeableOperationRequested(ctx context.Context, event *schemas.ChargeableOperationEvent) error {
	if r.feeErr != nil {
		return r.feeErr
	}
	r.fees = append(r.fees, event)
	return nil
}

func (r *recordJoinPublisher) PublishPlayerJoinedTournament(ctx context.Context, notice PlayerJoinedNotice) error {
	r.joined = append(r.joined, notice)
	return nil
}

func tournamentLobby() *entities.Lobby {
	tenant := uuid.New()
	client := uuid.New()
	owner := uuid.New()
	return &entities.Lobby{
		ID:         uuid.New(),
		Type:       entities.LobbyTypeTournament,
		Status:     entities.LobbyStatusOpen,
		GameID:     "cs2",
		Region:     "br",
		MaxPlayers: 4,
		TenantID:   tenant,
		ClientID:   client,
		CreatorID:  owner,
		ResourceOwner: common.ResourceOwner{
			TenantID: tenant,
			ClientID: client,
			UserID:   owner,
		},
		PrizePool: &entities.PrizePoolConfig{EntryFeeCents: 250},
	}
}

func TestJoinTournament_FeeThenPlayerJoined(t *testing.T) {
	lobby := tournamentLobby()
	store := &memLobbyStore{lobby: lobby}
	pub := &recordJoinPublisher{}
	uc := NewJoinTournamentUseCase(store, pub)
	player := uuid.New()
	got, slot, err := uc.Execute(context.Background(), JoinTournamentInput{
		LobbyID:       lobby.ID,
		PlayerID:      player,
		Caller:        lobby.ResourceOwner,
		CorrelationID: "corr-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if slot == nil || !got.HasPlayer(player) {
		t.Fatal("player not seated")
	}
	if len(pub.fees) != 1 || pub.fees[0].ChargeableOperationRequested.OperationType != schemas.OperationTypeTournamentEntry {
		t.Fatalf("fee events: %+v", pub.fees)
	}
	if len(pub.joined) != 1 || pub.joined[0].PlayerID != player || pub.joined[0].TenantID == "" || pub.joined[0].ResourceOwnerID == "" {
		t.Fatalf("joined notices: %+v", pub.joined)
	}
}

func TestJoinTournament_AlreadyJoinedPublishesNothing(t *testing.T) {
	lobby := tournamentLobby()
	player := uuid.New()
	lobby.PlayerSlots = []entities.PlayerSlot{{PlayerID: &player}}
	pub := &recordJoinPublisher{}
	uc := NewJoinTournamentUseCase(&memLobbyStore{lobby: lobby}, pub)
	_, _, err := uc.Execute(context.Background(), JoinTournamentInput{
		LobbyID:  lobby.ID,
		PlayerID: player,
		Caller:   lobby.ResourceOwner,
	})
	if !errors.Is(err, ErrAlreadyJoined) {
		t.Fatalf("got %v", err)
	}
	if len(pub.fees) != 0 || len(pub.joined) != 0 {
		t.Fatal("duplicate events")
	}
}

func TestJoinTournament_ClosedAndFullPublishNothing(t *testing.T) {
	lobby := tournamentLobby()
	lobby.Status = entities.LobbyStatusCancelled
	pub := &recordJoinPublisher{}
	uc := NewJoinTournamentUseCase(&memLobbyStore{lobby: lobby}, pub)
	_, _, err := uc.Execute(context.Background(), JoinTournamentInput{
		LobbyID:  lobby.ID,
		PlayerID: uuid.New(),
		Caller:   lobby.ResourceOwner,
	})
	if !errors.Is(err, ErrTournamentClosed) {
		t.Fatalf("got %v", err)
	}
	lobby.Status = entities.LobbyStatusOpen
	lobby.MaxPlayers = 1
	seated := uuid.New()
	lobby.PlayerSlots = []entities.PlayerSlot{{PlayerID: &seated}}
	_, _, err = uc.Execute(context.Background(), JoinTournamentInput{
		LobbyID:  lobby.ID,
		PlayerID: uuid.New(),
		Caller:   lobby.ResourceOwner,
	})
	if !errors.Is(err, ErrLobbyFull) {
		t.Fatalf("full: %v", err)
	}
	if len(pub.fees) != 0 || len(pub.joined) != 0 {
		t.Fatal("events on failure")
	}
}

func TestJoinTournament_FeeFailureDoesNotSeat(t *testing.T) {
	lobby := tournamentLobby()
	pub := &recordJoinPublisher{feeErr: errors.New("kafka down")}
	uc := NewJoinTournamentUseCase(&memLobbyStore{lobby: lobby}, pub)
	player := uuid.New()
	_, _, err := uc.Execute(context.Background(), JoinTournamentInput{
		LobbyID:  lobby.ID,
		PlayerID: player,
		Caller:   lobby.ResourceOwner,
	})
	if err == nil || lobby.HasPlayer(player) {
		t.Fatalf("err=%v seated=%v", err, lobby.HasPlayer(player))
	}
	if len(pub.joined) != 0 {
		t.Fatal("player joined must not publish when the fee fails")
	}
}

func TestJoinTournament_WrongTenant(t *testing.T) {
	lobby := tournamentLobby()
	uc := NewJoinTournamentUseCase(&memLobbyStore{lobby: lobby}, &recordJoinPublisher{})
	caller := lobby.ResourceOwner
	caller.TenantID = uuid.New()
	_, _, err := uc.Execute(context.Background(), JoinTournamentInput{
		LobbyID:  lobby.ID,
		PlayerID: uuid.New(),
		Caller:   caller,
	})
	if !errors.Is(err, ErrTournamentForbidden) {
		t.Fatalf("got %v", err)
	}
}
