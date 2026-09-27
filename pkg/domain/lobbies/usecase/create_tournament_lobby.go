package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// LobbyPersister stores a lobby.
type LobbyPersister interface {
	Create(ctx context.Context, lobby *entities.Lobby) error
}

// TournamentEventPublisher emits lobby, prize pool, and optional chargeable events.
// Match-making publishes only; wallet-api executes any creation fee.
type TournamentEventPublisher interface {
	PublishLobbyEvent(ctx context.Context, event *kafka.LobbyEvent) error
	PublishPrizePoolEvent(ctx context.Context, event *kafka.PrizePoolEvent) error
	PublishChargeableOperationRequested(ctx context.Context, event *schemas.ChargeableOperationEvent) error
}

// CreateTournamentLobbyCommand is the input for creating a tournament lobby and prize pool.
type CreateTournamentLobbyCommand struct {
	TournamentID     uuid.UUID
	Name             string
	GameID           string
	Region           string
	MaxPlayers       int
	DistributionRule string
	AmountCents      int64
	Currency         string
	CreationFeeCents int64
	Owner            common.ResourceOwner
}

// CreateTournamentLobbyUseCase creates a tournament lobby with a prize pool and publishes events.
type CreateTournamentLobbyUseCase struct {
	Lobbies LobbyPersister
	Events  TournamentEventPublisher
}

// Execute validates ownership, persists the lobby, and publishes LobbyCreated, PrizePoolCreated,
// and ChargeableOperationRequested when a creation fee applies.
func (uc *CreateTournamentLobbyUseCase) Execute(ctx context.Context, cmd CreateTournamentLobbyCommand) (*entities.Lobby, error) {
	if err := validateTournamentLobbyCommand(cmd); err != nil {
		return nil, err
	}
	if uc.Lobbies == nil || uc.Events == nil {
		return nil, fmt.Errorf("tournament lobby dependencies are not configured")
	}

	now := time.Now().UTC()
	lobbyID := uuid.New()
	poolID := uuid.New()
	currency := cmd.Currency
	if currency == "" {
		currency = "USD"
	}

	lobby := &entities.Lobby{
		ID:            lobbyID,
		TenantID:      cmd.Owner.TenantID,
		ClientID:      cmd.Owner.ClientID,
		CreatorID:     cmd.Owner.UserID,
		ResourceOwner: cmd.Owner,
		GameID:        cmd.GameID,
		Region:        cmd.Region,
		Name:          cmd.Name,
		Type:          entities.LobbyTypeTournament,
		Visibility:    entities.LobbyVisibilityPublic,
		MaxPlayers:    cmd.MaxPlayers,
		MinPlayers:    2,
		Status:        entities.LobbyStatusOpen,
		CreatedAt:     now,
		UpdatedAt:     now,
		ExpiresAt:     now.Add(24 * time.Hour),
		PlayerSlots:   []entities.PlayerSlot{},
		PrizePool: &entities.PrizePoolConfig{
			PrizePoolID:      poolID.String(),
			DistributionRule: cmd.DistributionRule,
			AmountCents:      cmd.AmountCents,
			Currency:         currency,
			Status:           entities.PrizePoolStatusOpen,
			TournamentID:     cmd.TournamentID.String(),
			CreationFeeCents: cmd.CreationFeeCents,
		},
		Metadata: map[string]interface{}{
			"tournament_id": cmd.TournamentID.String(),
		},
	}
	if err := lobby.ValidateOwnership(); err != nil {
		return nil, err
	}
	if err := uc.Lobbies.Create(ctx, lobby); err != nil {
		return nil, fmt.Errorf("persist lobby: %w", err)
	}

	ownerID := cmd.Owner.UserID.String()
	meta := map[string]string{
		"tenant_id":         cmd.Owner.TenantID.String(),
		"client_id":         cmd.Owner.ClientID.String(),
		"resource_owner_id": ownerID,
		"tournament_id":     cmd.TournamentID.String(),
		"prize_pool_id":     poolID.String(),
	}
	if err := uc.Events.PublishLobbyEvent(ctx, &kafka.LobbyEvent{
		LobbyID:   lobbyID,
		EventType: kafka.EventTypeLobbyCreated,
		GameType:  cmd.GameID,
		Region:    cmd.Region,
		Metadata:  meta,
	}); err != nil {
		return nil, fmt.Errorf("publish LobbyCreated: %w", err)
	}
	if err := uc.Events.PublishPrizePoolEvent(ctx, &kafka.PrizePoolEvent{
		PoolID:      poolID,
		LobbyID:     lobbyID,
		EventType:   kafka.EventTypePrizePoolCreated,
		TotalAmount: cmd.AmountCents,
		Currency:    currency,
		Metadata: map[string]string{
			"distribution_rule": cmd.DistributionRule,
			"status":            entities.PrizePoolStatusOpen,
			"tenant_id":         cmd.Owner.TenantID.String(),
			"client_id":         cmd.Owner.ClientID.String(),
			"resource_owner_id": ownerID,
			"tournament_id":     cmd.TournamentID.String(),
		},
	}); err != nil {
		return nil, fmt.Errorf("publish PrizePoolCreated: %w", err)
	}

	if cmd.CreationFeeCents > 0 {
		idem := fmt.Sprintf("%s:%s", schemas.OperationTypeLobbyCreationFee, lobbyID.String())
		if err := uc.Events.PublishChargeableOperationRequested(ctx, &schemas.ChargeableOperationEvent{
			ID:                uuid.New().String(),
			Type:              schemas.EventTypeChargeableOperationRequested,
			Source:            "match-making-api",
			SpecVersion:       schemas.CloudEventsSpecVersion,
			TimeUnixMs:        now.UnixMilli(),
			Subject:           ownerID,
			ResourceOwnerID:   ownerID,
			DataschemaVersion: schemas.SchemaVersionV1,
			ChargeableOperationRequested: &schemas.ChargeableOperationRequestedPayload{
				OperationType:   schemas.OperationTypeLobbyCreationFee,
				AmountCents:     cmd.CreationFeeCents,
				Currency:        currency,
				PlayerID:        ownerID,
				ResourceOwnerID: ownerID,
				TenantID:        cmd.Owner.TenantID.String(),
				ClientID:        cmd.Owner.ClientID.String(),
				GameID:          cmd.GameID,
				Region:          cmd.Region,
				IdempotencyKey:  idem,
				RequestedAtMs:   now.UnixMilli(),
			},
		}); err != nil {
			return nil, fmt.Errorf("publish ChargeableOperationRequested: %w", err)
		}
	}

	return lobby, nil
}

func validateTournamentLobbyCommand(cmd CreateTournamentLobbyCommand) error {
	if cmd.TournamentID == uuid.Nil {
		return fmt.Errorf("tournament_id is required")
	}
	if cmd.Name == "" || cmd.GameID == "" {
		return fmt.Errorf("name and game_id are required")
	}
	if cmd.MaxPlayers < 2 {
		return fmt.Errorf("max_players must be at least 2")
	}
	if cmd.AmountCents < 0 || cmd.CreationFeeCents < 0 {
		return fmt.Errorf("amounts must be non-negative")
	}
	if cmd.AmountCents > 0 && cmd.DistributionRule == "" {
		return fmt.Errorf("distribution_rule is required when prize pool amount is set")
	}
	if err := cmd.Owner.ValidateTenantClient(); err != nil {
		return err
	}
	if cmd.Owner.UserID == uuid.Nil {
		return fmt.Errorf("authenticated user is required")
	}
	return nil
}
