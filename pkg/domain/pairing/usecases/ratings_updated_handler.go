package usecases

import (
	"context"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// RatingsUpdatedHandler processes MatchResultsCalculated: computes rating deltas, persists, produces RatingsUpdated.
type RatingsUpdatedHandler struct {
	ratingRepo    pairing_out.PlayerRatingRepository
	processedStore pairing_out.RatingsProcessedStore
	publish       RatingsPublisher
	elo           *EloCalculator
}

// RatingsPublisher publishes RatingsUpdated to Kafka.
type RatingsPublisher interface {
	PublishRatingsUpdatedProto(ctx context.Context, event *schemas.MatchmakingEvent) error
}

// NewRatingsUpdatedHandler creates a handler for MatchResultsCalculated events.
func NewRatingsUpdatedHandler(
	ratingRepo pairing_out.PlayerRatingRepository,
	processedStore pairing_out.RatingsProcessedStore,
	publish RatingsPublisher,
	elo *EloCalculator,
) *RatingsUpdatedHandler {
	if elo == nil {
		elo = NewEloCalculator(EloKFactor)
	}
	return &RatingsUpdatedHandler{
		ratingRepo:     ratingRepo,
		processedStore: processedStore,
		publish:        publish,
		elo:            elo,
	}
}

// Handle processes a MatchResultsCalculated event: idempotent rating update and publish.
func (h *RatingsUpdatedHandler) Handle(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.MatchResultsCalculatedPayload) error {
	matchID, err := uuid.Parse(payload.GetMatchId())
	if err != nil {
		slog.ErrorContext(ctx, "Invalid match_id in MatchResultsCalculated",
			"match_id", payload.GetMatchId(),
			"error", err)
		return nil
	}

	// Idempotency: skip if already processed
	processed, err := h.processedStore.HasProcessed(ctx, matchID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to check ratings processed",
			"match_id", matchID,
			"error", err)
		return err
	}
	if processed {
		slog.InfoContext(ctx, "Ratings already processed for match, skipping",
			"match_id", matchID,
			"event_id", envelope.GetId())
		return nil
	}

	playerIDs := payload.GetPlayerIds()
	if len(playerIDs) == 0 {
		slog.WarnContext(ctx, "MatchResultsCalculated has no player_ids, skipping",
			"match_id", matchID)
		return nil
	}

	tenantID := strings.TrimSpace(payload.GetTenantId())
	clientID := strings.TrimSpace(payload.GetClientId())
	resourceOwnerID := strings.TrimSpace(payload.GetResourceOwnerId())
	gameID := strings.TrimSpace(payload.GetGameId())

	if tenantID == "" || clientID == "" || resourceOwnerID == "" {
		slog.ErrorContext(ctx, "MatchResultsCalculated missing tenant_id, client_id, or resource_owner_id",
			"match_id", matchID)
		return nil
	}

	// Determine winner: if winner_team_id equals a player_id (1v1 convention), that player won
	winnerPlayerID := ""
	if !payload.GetIsDraw() && payload.GetWinnerTeamId() != "" {
		for _, pid := range playerIDs {
			if pid == payload.GetWinnerTeamId() {
				winnerPlayerID = pid
				break
			}
		}
	}

	// Only support 2-player Elo for now; for other cases use simplified scoring
	deltas, err := h.computeDeltas(ctx, playerIDs, winnerPlayerID, payload.GetIsDraw(), tenantID, clientID, resourceOwnerID, gameID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to compute rating deltas",
			"match_id", matchID,
			"error", err)
		return err
	}

	if len(deltas) == 0 {
		slog.InfoContext(ctx, "No rating deltas to publish, skipping",
			"match_id", matchID)
		return nil
	}

	// Persist ratings, audit entries, and mark processed
	for _, d := range deltas {
		if err := h.ratingRepo.Save(ctx, d.NewRating); err != nil {
			slog.ErrorContext(ctx, "Failed to save player rating",
				"match_id", matchID,
				"player_id", d.PlayerID,
				"error", err)
			return err
		}
		entry := &entities.RatingAuditEntry{
			MatchID:          matchID,
			PlayerID:         d.PlayerID,
			MMRBefore:        d.MMRBefore,
			MMRAfter:         d.MMRAfter,
			Delta:            d.Delta,
			AlgorithmVersion: EloAlgorithmVersion,
			Reason:           "match_completion",
			UpdatedBy:        "match-making-api",
		}
		if err := h.ratingRepo.SaveAuditEntry(ctx, entry); err != nil {
			slog.WarnContext(ctx, "Failed to save audit entry (non-fatal)",
				"match_id", matchID,
				"player_id", d.PlayerID,
				"error", err)
		}
	}

	if err := h.processedStore.MarkProcessed(ctx, matchID); err != nil {
		slog.ErrorContext(ctx, "Failed to mark ratings processed",
			"match_id", matchID,
			"error", err)
		return err
	}

	// Produce RatingsUpdated
	updatedAt := time.Now().UTC()
	updatedAtMs := updatedAt.UnixMilli()

	protoDeltas := make([]*schemas.PlayerRatingDelta, len(deltas))
	for i, d := range deltas {
		protoDeltas[i] = &schemas.PlayerRatingDelta{
			PlayerId:  d.PlayerID,
			MmrBefore: d.MMRBefore,
			MmrAfter:  d.MMRAfter,
			Delta:     d.Delta,
		}
	}

	event := &schemas.MatchmakingEvent{
		Envelope: &schemas.EventEnvelope{
			Id:                uuid.New().String(),
			Type:              schemas.EventTypeRatingsUpdated,
			Source:            "match-making-api",
			Specversion:       schemas.CloudEventsSpecVersion,
			Time:              timestamppb.New(updatedAt),
			Subject:           matchID.String(),
			ResourceOwnerId:   resourceOwnerID,
			CorrelationId:     envelope.GetCorrelationId(),
			DataschemaVersion: schemas.SchemaVersionV1,
		},
		Data: &schemas.MatchmakingEvent_RatingsUpdated{
			RatingsUpdated: &schemas.RatingsUpdatedPayload{
				MatchId:          matchID.String(),
				Deltas:           protoDeltas,
				UpdatedAtEpochMs: updatedAtMs,
				TenantId:         tenantID,
				ClientId:         clientID,
				ResourceOwnerId:  resourceOwnerID,
				GameId:           gameID,
				AuditTrail: &schemas.RatingAuditTrail{
					AlgorithmVersion: EloAlgorithmVersion,
					Reason:           "match_completion",
					UpdatedBy:        "match-making-api",
				},
			},
		},
	}

	if err := h.publish.PublishRatingsUpdatedProto(ctx, event); err != nil {
		slog.ErrorContext(ctx, "Failed to publish RatingsUpdated",
			"match_id", matchID,
			"error", err)
		return err
	}

	slog.InfoContext(ctx, "RatingsUpdated published",
		"match_id", matchID,
		"player_count", len(deltas),
		"event_id", envelope.GetId())

	return nil
}

type ratingDelta struct {
	PlayerID  string
	MMRBefore int32
	MMRAfter  int32
	Delta     int32
	NewRating *entities.PlayerRating
}

func (h *RatingsUpdatedHandler) computeDeltas(
	ctx context.Context,
	playerIDs []string,
	winnerPlayerID string,
	isDraw bool,
	tenantID, clientID, resourceOwnerID, gameID string,
) ([]ratingDelta, error) {
	// Fetch current ratings
	ratings := make([]int32, len(playerIDs))
	for i, pid := range playerIDs {
		pr, err := h.ratingRepo.GetByPlayer(ctx, pid, gameID, tenantID, clientID)
		if err != nil {
			return nil, err
		}
		if pr != nil {
			ratings[i] = pr.MMR
		} else {
			ratings[i] = int32(DefaultMMR)
		}
	}

	result := make([]ratingDelta, len(playerIDs))

	if len(playerIDs) == 2 {
		// Standard 1v1 Elo
		actual0 := 0.5
		actual1 := 0.5
		if !isDraw {
			if playerIDs[0] == winnerPlayerID {
				actual0, actual1 = 1, 0
			} else if playerIDs[1] == winnerPlayerID {
				actual0, actual1 = 0, 1
			}
		}

		delta0 := h.elo.Delta(ratings[0], ratings[1], actual0)
		delta1 := h.elo.Delta(ratings[1], ratings[0], actual1)

		newMMR0 := ratings[0] + delta0
		newMMR1 := ratings[1] + delta1

		result[0] = ratingDelta{
			PlayerID:  playerIDs[0],
			MMRBefore: ratings[0],
			MMRAfter:  newMMR0,
			Delta:     delta0,
			NewRating: &entities.PlayerRating{
				PlayerID:        playerIDs[0],
				GameID:          gameID,
				TenantID:        tenantID,
				ClientID:        clientID,
				ResourceOwnerID: resourceOwnerID,
				MMR:             newMMR0,
			},
		}
		result[1] = ratingDelta{
			PlayerID:  playerIDs[1],
			MMRBefore: ratings[1],
			MMRAfter:  newMMR1,
			Delta:     delta1,
			NewRating: &entities.PlayerRating{
				PlayerID:        playerIDs[1],
				GameID:          gameID,
				TenantID:        tenantID,
				ClientID:        clientID,
				ResourceOwnerID: resourceOwnerID,
				MMR:             newMMR1,
			},
		}
		return result, nil
	}

	// Multi-player: simplified model - use average opponent rating
	if len(playerIDs) <= 1 {
		return nil, nil
	}
	// Each player gets expected 1/N, actual 1 if winner else 0 (or 1/N if draw)
	n := float64(len(playerIDs))
	expected := 1.0 / n
	for i, pid := range playerIDs {
		actual := expected
		if !isDraw && pid == winnerPlayerID {
			actual = 1
		}
		opponentAvg := int32(0)
		for j, r := range ratings {
			if j != i {
				opponentAvg += r
			}
		}
		if len(playerIDs) > 1 {
			opponentAvg /= int32(len(playerIDs) - 1)
		}
		delta := h.elo.Delta(ratings[i], opponentAvg, actual)
		newMMR := ratings[i] + delta
		newMMR = int32(math.Max(0, float64(newMMR)))
		result[i] = ratingDelta{
			PlayerID:  pid,
			MMRBefore: ratings[i],
			MMRAfter:  newMMR,
			Delta:     delta,
			NewRating: &entities.PlayerRating{
				PlayerID:        pid,
				GameID:          gameID,
				TenantID:        tenantID,
				ClientID:        clientID,
				ResourceOwnerID: resourceOwnerID,
				MMR:             newMMR,
			},
		}
	}
	return result, nil
}
