package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// MatchCompletedHandler processes MatchCompleted events: calculates results, persists, produces MatchResultsCalculated.
type MatchCompletedHandler struct {
	repo    pairing_out.MatchResultRepository
	publish MatchResultsPublisher
}

// MatchResultsPublisher publishes MatchResultsCalculated to Kafka.
type MatchResultsPublisher interface {
	PublishMatchResultsCalculatedProto(ctx context.Context, event *schemas.MatchmakingEvent) error
}

// NewMatchCompletedHandler creates a handler for MatchCompleted events.
func NewMatchCompletedHandler(repo pairing_out.MatchResultRepository, publish MatchResultsPublisher) *MatchCompletedHandler {
	return &MatchCompletedHandler{
		repo:    repo,
		publish: publish,
	}
}

// Handle processes a MatchCompleted event: idempotent calculate, persist, produce.
func (h *MatchCompletedHandler) Handle(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.MatchCompletedPayload) error {
	matchID, err := uuid.Parse(payload.GetMatchId())
	if err != nil {
		slog.ErrorContext(ctx, "Invalid match_id in MatchCompleted",
			"match_id", payload.GetMatchId(),
			"error", err)
		return nil // Don't retry — invalid data
	}

	// Idempotency: if results already exist, skip (don't double-count)
	existing, err := h.repo.GetByMatchID(ctx, matchID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to check existing match result",
			"match_id", matchID,
			"error", err)
		return err
	}
	if existing != nil {
		slog.InfoContext(ctx, "Match result already exists, skipping (idempotent)",
			"match_id", matchID,
			"event_id", envelope.GetId())
		return nil
	}

	calculatedAt := time.Now().UTC()
	calculatedAtMs := calculatedAt.UnixMilli()

	// Persist with resource ownership
	result := &entities.MatchResult{
		MatchID:         matchID,
		PlayerIDs:       payload.GetPlayerIds(),
		WinnerTeamID:   payload.GetWinnerTeamId(),
		IsDraw:         payload.GetIsDraw(),
		CompletedAtMs:  payload.GetCompletedAtEpochMs(),
		CalculatedAtMs: calculatedAtMs,
		TenantID:       payload.GetTenantId(),
		ClientID:       payload.GetClientId(),
		ResourceOwnerID: envelope.GetResourceOwnerId(),
		SourceEventID:   envelope.GetId(),
		CalculatedAt:    calculatedAt,
	}

	saved, err := h.repo.Save(ctx, result)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to persist match result",
			"match_id", matchID,
			"error", err)
		return err
	}

	// Produce MatchResultsCalculated
	matchResultsEvent := &schemas.MatchmakingEvent{
		Envelope: &schemas.EventEnvelope{
			Id:                uuid.New().String(),
			Type:              schemas.EventTypeMatchResultsCalculated,
			Source:             "match-making-api",
			Specversion:        schemas.CloudEventsSpecVersion,
			Time:               timestamppb.New(calculatedAt),
			Subject:            matchID.String(),
			ResourceOwnerId:    envelope.GetResourceOwnerId(),
			CorrelationId:      envelope.GetCorrelationId(),
			DataschemaVersion:  schemas.SchemaVersionV1,
		},
		Data: &schemas.MatchmakingEvent_MatchResultsCalculated{
			MatchResultsCalculated: &schemas.MatchResultsCalculatedPayload{
				MatchId:             saved.MatchID.String(),
				PlayerIds:           saved.PlayerIDs,
				WinnerTeamId:        saved.WinnerTeamID,
				IsDraw:              saved.IsDraw,
				CompletedAtEpochMs: saved.CompletedAtMs,
				CalculatedAtEpochMs: saved.CalculatedAtMs,
				TenantId:            saved.TenantID,
				ClientId:            saved.ClientID,
				ResourceOwnerId:     saved.ResourceOwnerID,
				LobbyId:             payload.LobbyId,
				PrizePoolId:         payload.PrizePoolId,
			},
		},
	}

	if err := h.publish.PublishMatchResultsCalculatedProto(ctx, matchResultsEvent); err != nil {
		slog.ErrorContext(ctx, "Failed to publish MatchResultsCalculated",
			"match_id", matchID,
			"event_id", envelope.GetId(),
			"error", err)
		return err
	}

	slog.InfoContext(ctx, "MatchResultsCalculated published",
		"match_id", matchID,
		"winner_team_id", saved.WinnerTeamID,
		"is_draw", saved.IsDraw,
		"event_id", envelope.GetId())

	return nil
}
