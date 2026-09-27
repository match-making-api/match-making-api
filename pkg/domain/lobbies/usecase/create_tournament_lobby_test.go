package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

type memLobbies struct {
	saved *entities.Lobby
	err   error
}

func (m *memLobbies) Create(_ context.Context, lobby *entities.Lobby) error {
	if m.err != nil {
		return m.err
	}
	m.saved = lobby
	return nil
}

type memEvents struct {
	lobby   *kafka.LobbyEvent
	pool    *kafka.PrizePoolEvent
	charge  *schemas.ChargeableOperationEvent
}

func (m *memEvents) PublishLobbyEvent(_ context.Context, event *kafka.LobbyEvent) error {
	m.lobby = event
	return nil
}
func (m *memEvents) PublishPrizePoolEvent(_ context.Context, event *kafka.PrizePoolEvent) error {
	m.pool = event
	return nil
}
func (m *memEvents) PublishChargeableOperationRequested(_ context.Context, event *schemas.ChargeableOperationEvent) error {
	m.charge = event
	return nil
}

func owner() common.ResourceOwner {
	return common.ResourceOwner{
		TenantID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		ClientID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		UserID:   uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
	}
}

func TestCreateTournamentLobby_PublishesEventsWithoutFee(t *testing.T) {
	lobbies := &memLobbies{}
	events := &memEvents{}
	uc := &CreateTournamentLobbyUseCase{Lobbies: lobbies, Events: events}
	lobby, err := uc.Execute(context.Background(), CreateTournamentLobbyCommand{
		TournamentID:     uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
		Name:             "Cup",
		GameID:           "cs2",
		Region:           "br",
		MaxPlayers:       8,
		DistributionRule: "winner_takes_all",
		AmountCents:      10000,
		Owner:            owner(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if lobby.Type != entities.LobbyTypeTournament || lobby.PrizePool.Status != entities.PrizePoolStatusOpen {
		t.Fatalf("lobby: %+v", lobby.PrizePool)
	}
	if lobby.PrizePool.AmountCents != 10000 || lobby.PrizePool.DistributionRule != "winner_takes_all" {
		t.Fatalf("pool: %+v", lobby.PrizePool)
	}
	if events.lobby == nil || events.lobby.EventType != kafka.EventTypeLobbyCreated {
		t.Fatalf("lobby event: %+v", events.lobby)
	}
	if events.pool == nil || events.pool.EventType != kafka.EventTypePrizePoolCreated || events.pool.TotalAmount != 10000 {
		t.Fatalf("pool event: %+v", events.pool)
	}
	if events.charge != nil {
		t.Fatal("free creation must not publish chargeable event")
	}
}

func TestCreateTournamentLobby_PublishesCreationFee(t *testing.T) {
	events := &memEvents{}
	uc := &CreateTournamentLobbyUseCase{Lobbies: &memLobbies{}, Events: events}
	_, err := uc.Execute(context.Background(), CreateTournamentLobbyCommand{
		TournamentID:     uuid.New(),
		Name:             "Cup",
		GameID:           "cs2",
		MaxPlayers:       4,
		DistributionRule: "top_3",
		AmountCents:      500,
		CreationFeeCents: 250,
		Owner:            owner(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if events.charge == nil || events.charge.ChargeableOperationRequested == nil {
		t.Fatal("expected chargeable event")
	}
	p := events.charge.ChargeableOperationRequested
	if p.OperationType != schemas.OperationTypeLobbyCreationFee || p.AmountCents != 250 {
		t.Fatalf("charge: %+v", p)
	}
	if p.IdempotencyKey == "" || p.TenantID == "" {
		t.Fatalf("missing ownership/idempotency: %+v", p)
	}
}

func TestCreateTournamentLobby_RejectsMissingOwner(t *testing.T) {
	uc := &CreateTournamentLobbyUseCase{Lobbies: &memLobbies{}, Events: &memEvents{}}
	_, err := uc.Execute(context.Background(), CreateTournamentLobbyCommand{
		TournamentID: uuid.New(),
		Name:         "Cup",
		GameID:       "cs2",
		MaxPlayers:   4,
		AmountCents:  0,
		Owner:        common.ResourceOwner{},
	})
	if !errors.Is(err, common.ErrMissingResourceOwnership) {
		t.Fatalf("want ownership error, got %v", err)
	}
}
