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

// AnalyticsTrackedHandler processes MatchResultsCalculated: produces AnalyticsTracked for analytics pipeline (#32).
type AnalyticsTrackedHandler struct {
	trackedStore pairing_out.AnalyticsTrackedStore
	publish      AnalyticsTrackedPublisher
}

// AnalyticsTrackedPublisher publishes AnalyticsTracked to Kafka.
type AnalyticsTrackedPublisher interface {
	PublishAnalyticsTrackedProto(ctx context.Context, event *schemas.MatchmakingEvent) error
}

// NewAnalyticsTrackedHandler creates a handler for analytics events.
func NewAnalyticsTrackedHandler(
	trackedStore pairing_out.AnalyticsTrackedStore,
	publish AnalyticsTrackedPublisher,
) *AnalyticsTrackedHandler {
	return &AnalyticsTrackedHandler{
		trackedStore: trackedStore,
		publish:      publish,
	}
}

// Handle processes a MatchResultsCalculated event: idempotent analytics publication.
func (h *AnalyticsTrackedHandler) Handle(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.MatchResultsCalculatedPayload) error {
	matchID, err := uuid.Parse(payload.GetMatchId())
	if err != nil {
		slog.ErrorContext(ctx, "Invalid match_id in MatchResultsCalculated (analytics)",
			"match_id", payload.GetMatchId(),
			"error", err)
		return nil
	}

	// Idempotency: skip if already tracked
	tracked, err := h.trackedStore.HasTracked(ctx, matchID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to check analytics tracked",
			"match_id", matchID,
			"error", err)
		return err
	}
	if tracked {
		slog.InfoContext(ctx, "Analytics already tracked for match, skipping",
			"match_id", matchID,
			"event_id", envelope.GetId())
		return nil
	}

	tenantID := strings.TrimSpace(payload.GetTenantId())
	clientID := strings.TrimSpace(payload.GetClientId())
	resourceOwnerID := strings.TrimSpace(payload.GetResourceOwnerId())
	if tenantID == "" || clientID == "" || resourceOwnerID == "" {
		slog.ErrorContext(ctx, "MatchResultsCalculated missing tenant_id, client_id, or resource_owner_id (analytics)",
			"match_id", matchID)
		return nil
	}

	trackedAt := time.Now().UTC()
	trackedAtMs := trackedAt.UnixMilli()

	// Duration: 0 (MatchResultsCalculated does not have start time; analytics can join from MatchStarted if needed)
	durationMs := int64(0)

	event := &schemas.MatchmakingEvent{
		Envelope: &schemas.EventEnvelope{
			Id:                uuid.New().String(),
			Type:              schemas.EventTypeAnalyticsTracked,
			Source:             "match-making-api",
			Specversion:        schemas.CloudEventsSpecVersion,
			Time:               timestamppb.New(trackedAt),
			Subject:            matchID.String(),
			ResourceOwnerId:   resourceOwnerID,
			CorrelationId:     envelope.GetCorrelationId(),
			DataschemaVersion: schemas.SchemaVersionV1,
		},
		Data: &schemas.MatchmakingEvent_AnalyticsTracked{
			AnalyticsTracked: &schemas.AnalyticsTrackedPayload{
				MatchId:             matchID.String(),
				PlayerIds:           payload.GetPlayerIds(),
				WinnerTeamId:        payload.GetWinnerTeamId(),
				IsDraw:              payload.GetIsDraw(),
				CompletedAtEpochMs:  payload.GetCompletedAtEpochMs(),
				TrackedAtEpochMs:    trackedAtMs,
				DurationMs:          durationMs,
				TenantId:            tenantID,
				ClientId:            clientID,
				ResourceOwnerId:     resourceOwnerID,
				GameId:              payload.GetGameId(),
				LobbyId:             payload.GetLobbyId(),
				PrizePoolId:         payload.GetPrizePoolId(),
			},
		},
	}

	if err := h.publish.PublishAnalyticsTrackedProto(ctx, event); err != nil {
		slog.ErrorContext(ctx, "Failed to publish AnalyticsTracked",
			"match_id", matchID,
			"error", err)
		return err
	}

	if err := h.trackedStore.MarkTracked(ctx, matchID); err != nil {
		slog.ErrorContext(ctx, "Failed to mark analytics tracked (event already published)",
			"match_id", matchID,
			"error", err)
		return err
	}

	slog.InfoContext(ctx, "AnalyticsTracked published",
		"match_id", matchID,
		"event_id", envelope.GetId())

	return nil
}
