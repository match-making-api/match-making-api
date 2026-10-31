package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

var (
	// ErrLobbyFull is returned when the roster has no free slot.
	ErrLobbyFull = errors.New("lobby full")
	// ErrAlreadyJoined is returned when the player is already seated. No events are published.
	ErrAlreadyJoined = errors.New("already joined")
)

// PlayerJoinedNotice is the PlayerJoinedTournament payload on matchmaking.lobby.events.
type PlayerJoinedNotice struct {
	LobbyID         uuid.UUID
	PlayerID        uuid.UUID
	GameType        string
	Region          string
	Status          string
	TenantID        string
	ClientID        string
	ResourceOwnerID string
}

// TournamentJoinPublisher emits the entry-fee event and PlayerJoinedTournament.
// It must not charge a wallet.
type TournamentJoinPublisher interface {
	PublishChargeableOperationRequested(ctx context.Context, event *schemas.ChargeableOperationEvent) error
	PublishPlayerJoinedTournament(ctx context.Context, notice PlayerJoinedNotice) error
}

// JoinTournamentInput is the authenticated join request.
type JoinTournamentInput struct {
	LobbyID       uuid.UUID
	PlayerID      uuid.UUID
	Caller        common.ResourceOwner
	PlayerMMR     int
	PlayerRank    string
	CorrelationID string
}

// JoinTournamentUseCase seats a player and publishes PlayerJoinedTournament
// after the entry-fee event when a fee applies.
type JoinTournamentUseCase struct {
	store     TournamentLobbyStore
	publisher TournamentJoinPublisher
}

// NewJoinTournamentUseCase builds the join use case.
func NewJoinTournamentUseCase(store TournamentLobbyStore, publisher TournamentJoinPublisher) *JoinTournamentUseCase {
	return &JoinTournamentUseCase{store: store, publisher: publisher}
}

// Execute validates, publishes the entry fee when required, saves the seat, then publishes PlayerJoinedTournament.
func (u *JoinTournamentUseCase) Execute(ctx context.Context, in JoinTournamentInput) (*entities.Lobby, *entities.PlayerSlot, error) {
	if u == nil || u.store == nil {
		return nil, nil, errors.New("join tournament store is nil")
	}
	if in.Caller.TenantID == uuid.Nil || in.Caller.ClientID == uuid.Nil {
		return nil, nil, common.ErrMissingResourceOwnership
	}

	lobby, err := u.store.GetByID(ctx, in.LobbyID)
	if err != nil {
		return nil, nil, err
	}
	if lobby == nil {
		return nil, nil, errors.New("lobby not found")
	}
	if err := lobby.ValidateOwnership(); err != nil {
		return nil, nil, err
	}
	if lobby.ResourceOwner.TenantID != in.Caller.TenantID || lobby.ResourceOwner.ClientID != in.Caller.ClientID {
		return nil, nil, ErrTournamentForbidden
	}
	if lobby.Type != entities.LobbyTypeTournament {
		return nil, nil, ErrNotATournament
	}
	if lobby.HasPlayer(in.PlayerID) {
		return lobby, nil, ErrAlreadyJoined
	}
	if lobby.Status != entities.LobbyStatusOpen {
		return nil, nil, ErrTournamentClosed
	}
	if lobby.IsFull() {
		return nil, nil, ErrLobbyFull
	}

	feeInput := EntryFeeInputFromLobby(lobby, in.PlayerID, in.CorrelationID)
	feeEvent, hasFee := BuildTournamentEntryFeeEvent(feeInput)
	if hasFee {
		if u.publisher == nil {
			return nil, nil, errors.New("entry fee publisher is nil")
		}
		if err := u.publisher.PublishChargeableOperationRequested(ctx, feeEvent); err != nil {
			return nil, nil, err
		}
	}

	slotNumber := len(lobby.PlayerSlots) + 1
	playerID := in.PlayerID
	slot := entities.PlayerSlot{
		SlotNumber: slotNumber,
		PlayerID:   &playerID,
		IsReady:    false,
		JoinedAt:   time.Now().UTC(),
		MMR:        in.PlayerMMR,
		Rank:       in.PlayerRank,
		Team:       (slotNumber % 2) + 1,
	}
	lobby.PlayerSlots = append(lobby.PlayerSlots, slot)
	lobby.UpdatedAt = time.Now().UTC()
	if err := u.store.Update(ctx, lobby); err != nil {
		return nil, nil, err
	}

	if u.publisher != nil {
		if err := u.publisher.PublishPlayerJoinedTournament(ctx, playerJoinedNotice(lobby, in.PlayerID)); err != nil {
			return lobby, &slot, err
		}
	}
	return lobby, &slot, nil
}

func playerJoinedNotice(lobby *entities.Lobby, playerID uuid.UUID) PlayerJoinedNotice {
	return PlayerJoinedNotice{
		LobbyID:         lobby.ID,
		PlayerID:        playerID,
		GameType:        lobby.GameID,
		Region:          lobby.Region,
		Status:          string(lobby.Status),
		TenantID:        lobby.ResourceOwner.TenantID.String(),
		ClientID:        lobby.ResourceOwner.ClientID.String(),
		ResourceOwnerID: lobby.ResourceOwner.UserID.String(),
	}
}
