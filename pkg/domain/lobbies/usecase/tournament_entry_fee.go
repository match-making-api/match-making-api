package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// TournamentEntryFeeInput is the data needed to publish a tournament entry fee.
// Match-making does not charge the fee. Wallet API consumes the event.
type TournamentEntryFeeInput struct {
	TournamentID    string
	PlayerID        string
	AmountCents     int64
	Currency        string
	ResourceOwnerID string
	TenantID        string
	ClientID        string
	GameID          string
	Region          string
	CorrelationID   string
}

// BuildTournamentEntryFeeEvent returns a ChargeableOperationRequested when the entry fee is positive.
// A zero or negative fee returns ok=false and no event.
func BuildTournamentEntryFeeEvent(in TournamentEntryFeeInput) (*schemas.ChargeableOperationEvent, bool) {
	if in.AmountCents <= 0 || strings.TrimSpace(in.PlayerID) == "" || strings.TrimSpace(in.TournamentID) == "" {
		return nil, false
	}
	now := time.Now().UTC()
	currency := in.Currency
	if currency == "" {
		currency = "USD"
	}
	idem := fmt.Sprintf("%s:%s:%s", schemas.OperationTypeTournamentEntry, in.PlayerID, in.TournamentID)
	return &schemas.ChargeableOperationEvent{
		ID:                uuid.New().String(),
		Type:              schemas.EventTypeChargeableOperationRequested,
		Source:            "match-making-api",
		SpecVersion:       schemas.CloudEventsSpecVersion,
		TimeUnixMs:        now.UnixMilli(),
		Subject:           in.PlayerID,
		ResourceOwnerID:   in.ResourceOwnerID,
		CorrelationID:     in.CorrelationID,
		DataschemaVersion: schemas.SchemaVersionV1,
		ChargeableOperationRequested: &schemas.ChargeableOperationRequestedPayload{
			OperationType:   schemas.OperationTypeTournamentEntry,
			AmountCents:     in.AmountCents,
			Currency:        currency,
			PlayerID:        in.PlayerID,
			ResourceOwnerID: in.ResourceOwnerID,
			TenantID:        in.TenantID,
			ClientID:        in.ClientID,
			GameID:          in.GameID,
			Region:          in.Region,
			TournamentID:    in.TournamentID,
			CorrelationID:   in.CorrelationID,
			IdempotencyKey:  idem,
			RequestedAtMs:   now.UnixMilli(),
		},
	}, true
}

// EntryFeeInputFromLobby maps a tournament lobby join into a fee event input.
// Non-tournament lobbies and lobbies without a positive entry fee produce a zero amount.
func EntryFeeInputFromLobby(lobby *entities.Lobby, playerID uuid.UUID, correlationID string) TournamentEntryFeeInput {
	in := TournamentEntryFeeInput{
		PlayerID:      playerID.String(),
		CorrelationID: correlationID,
		Currency:      "USD",
	}
	if lobby == nil {
		return in
	}
	in.TournamentID = lobby.ID.String()
	in.GameID = lobby.GameID
	in.Region = lobby.Region
	lobby.EnsureResourceOwner()
	if lobby.ResourceOwner.UserID != uuid.Nil {
		in.ResourceOwnerID = lobby.ResourceOwner.UserID.String()
	} else if lobby.CreatorID != uuid.Nil {
		in.ResourceOwnerID = lobby.CreatorID.String()
	}
	if lobby.TenantID != uuid.Nil {
		in.TenantID = lobby.TenantID.String()
	} else if lobby.ResourceOwner.TenantID != uuid.Nil {
		in.TenantID = lobby.ResourceOwner.TenantID.String()
	}
	if lobby.ClientID != uuid.Nil {
		in.ClientID = lobby.ClientID.String()
	} else if lobby.ResourceOwner.ClientID != uuid.Nil {
		in.ClientID = lobby.ResourceOwner.ClientID.String()
	}
	if lobby.Type == entities.LobbyTypeTournament && lobby.PrizePool != nil && lobby.PrizePool.EntryFeeCents > 0 {
		in.AmountCents = int64(lobby.PrizePool.EntryFeeCents)
	}
	return in
}
