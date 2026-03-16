package usecases

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// PrizeDistributionHandler processes MatchResultsCalculated: determines winners, resolves prize amounts, produces PrizeDistributed.
// Match-making does NOT execute wallet transfers; Wallet API consumes the event and executes transfers.
type PrizeDistributionHandler struct {
	prizeResolver   pairing_out.PrizeAmountResolver
	distributedStore pairing_out.PrizesDistributedStore
	publish         PrizeDistributedPublisher
}

// PrizeDistributedPublisher publishes PrizeDistributed to Kafka.
type PrizeDistributedPublisher interface {
	PublishPrizeDistributedProto(ctx context.Context, event *schemas.MatchmakingEvent) error
}

// NewPrizeDistributionHandler creates a handler for prize distribution events.
func NewPrizeDistributionHandler(
	prizeResolver pairing_out.PrizeAmountResolver,
	distributedStore pairing_out.PrizesDistributedStore,
	publish PrizeDistributedPublisher,
) *PrizeDistributionHandler {
	return &PrizeDistributionHandler{
		prizeResolver:   prizeResolver,
		distributedStore: distributedStore,
		publish:         publish,
	}
}

// Handle processes a MatchResultsCalculated event: idempotent prize distribution and publish.
func (h *PrizeDistributionHandler) Handle(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.MatchResultsCalculatedPayload) error {
	matchID, err := uuid.Parse(payload.GetMatchId())
	if err != nil {
		slog.ErrorContext(ctx, "Invalid match_id in MatchResultsCalculated (prize)",
			"match_id", payload.GetMatchId(),
			"error", err)
		return nil
	}

	// Idempotency: skip if already distributed
	distributed, err := h.distributedStore.HasDistributed(ctx, matchID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to check prizes distributed",
			"match_id", matchID,
			"error", err)
		return err
	}
	if distributed {
		slog.InfoContext(ctx, "Prizes already distributed for match, skipping",
			"match_id", matchID,
			"event_id", envelope.GetId())
		return nil
	}

	// Determine winners: same convention as ratings (winner_team_id in player_ids for 1v1)
	winnerPlayerIDs := h.getWinnerPlayerIDs(payload)
	if len(winnerPlayerIDs) == 0 {
		slog.InfoContext(ctx, "No winners for match (draw or unknown), skipping prize distribution",
			"match_id", matchID)
		return nil
	}

	tenantID := strings.TrimSpace(payload.GetTenantId())
	clientID := strings.TrimSpace(payload.GetClientId())
	resourceOwnerID := strings.TrimSpace(payload.GetResourceOwnerId())
	if tenantID == "" || clientID == "" || resourceOwnerID == "" {
		slog.ErrorContext(ctx, "MatchResultsCalculated missing tenant_id, client_id, or resource_owner_id (prize)",
			"match_id", matchID)
		return nil
	}

	// Resolve prize amounts
	req := pairing_out.PrizeResolveRequest{
		MatchID:         matchID.String(),
		WinnerPlayerIDs: winnerPlayerIDs,
		LobbyID:         payload.GetLobbyId(),
		PrizePoolID:     payload.GetPrizePoolId(),
		TenantID:        tenantID,
		ClientID:        clientID,
	}

	prizes, err := h.prizeResolver.Resolve(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to resolve prize amounts",
			"match_id", matchID,
			"error", err)
		return err
	}

	if len(prizes) == 0 {
		slog.InfoContext(ctx, "No prize amounts resolved for match, skipping",
			"match_id", matchID,
			"lobby_id", req.LobbyID,
			"prize_pool_id", req.PrizePoolID)
		return nil
	}

	// Mark distributed before publish (optimistic - if publish fails we retry and HasDistributed will be false)
	// Actually we should mark AFTER successful publish to avoid "distributed but not sent". Let me mark after publish.
	distributedAt := time.Now().UTC()
	distributedAtMs := distributedAt.UnixMilli()

	winnerDetails := make([]*schemas.WinnerPrizeDetail, len(prizes))
	currency := "USD"
	for i, p := range prizes {
		c := p.Currency
		if c == "" {
			c = currency
		}
		winnerDetails[i] = &schemas.WinnerPrizeDetail{
			PlayerId:    p.PlayerID,
			AmountCents: p.AmountCents,
			Currency:    c,
		}
	}

	event := &schemas.MatchmakingEvent{
		Envelope: &schemas.EventEnvelope{
			Id:                uuid.New().String(),
			Type:              schemas.EventTypePrizeDistributed,
			Source:             "match-making-api",
			Specversion:        schemas.CloudEventsSpecVersion,
			Time:               timestamppb.New(distributedAt),
			Subject:            matchID.String(),
			ResourceOwnerId:   resourceOwnerID,
			CorrelationId:     envelope.GetCorrelationId(),
			DataschemaVersion: schemas.SchemaVersionV1,
		},
		Data: &schemas.MatchmakingEvent_PrizeDistributed{
			PrizeDistributed: &schemas.PrizeDistributedPayload{
				MatchId:              matchID.String(),
				WinnerDetails:        winnerDetails,
				DistributedAtEpochMs: distributedAtMs,
				TenantId:             tenantID,
				ClientId:             clientID,
				ResourceOwnerId:     resourceOwnerID,
				LobbyId:             payload.GetLobbyId(),
				PrizePoolId:         payload.GetPrizePoolId(),
				Currency:            currency,
			},
		},
	}

	if err := h.publish.PublishPrizeDistributedProto(ctx, event); err != nil {
		slog.ErrorContext(ctx, "Failed to publish PrizeDistributed",
			"match_id", matchID,
			"error", err)
		return err
	}

	if err := h.distributedStore.MarkDistributed(ctx, matchID); err != nil {
		slog.ErrorContext(ctx, "Failed to mark prizes distributed (event already published)",
			"match_id", matchID,
			"error", err)
		return err
	}

	slog.InfoContext(ctx, "PrizeDistributed published",
		"match_id", matchID,
		"winner_count", len(winnerDetails),
		"event_id", envelope.GetId())

	return nil
}

// getWinnerPlayerIDs returns player IDs that won. For 1v1: winner_team_id in player_ids.
// For draw: empty. For team games without player mapping: empty (no prizes without mapping).
func (h *PrizeDistributionHandler) getWinnerPlayerIDs(payload *schemas.MatchResultsCalculatedPayload) []string {
	if payload.GetIsDraw() {
		return nil
	}
	winnerTeamID := payload.GetWinnerTeamId()
	if winnerTeamID == "" {
		return nil
	}
	playerIDs := payload.GetPlayerIds()
	for _, pid := range playerIDs {
		if pid == winnerTeamID {
			return []string{pid}
		}
	}
	// winner_team_id is a team ID, not player ID - would need team mapping
	return nil
}
