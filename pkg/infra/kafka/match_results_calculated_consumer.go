package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	kafkago "github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

// RatingsUpdatedHandler is the domain-level function that processes a validated MatchResultsCalculated event.
type RatingsUpdatedHandler func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.MatchResultsCalculatedPayload) error

// MatchResultsCalculatedConsumer consumes MatchResultsCalculated events from matchmaking.matches.results.
// Validates resource ownership and delegates to the domain handler for rating updates.
type MatchResultsCalculatedConsumer struct {
	consumer *Consumer
	handler  RatingsUpdatedHandler
}

// NewMatchResultsCalculatedConsumer creates a consumer for the matchmaking.matches.results topic.
func NewMatchResultsCalculatedConsumer(client *Client, groupID string, handler RatingsUpdatedHandler) *MatchResultsCalculatedConsumer {
	config := DefaultConsumerConfig(groupID, []string{TopicMatchesResults})
	consumer := NewConsumer(client, config)

	mcc := &MatchResultsCalculatedConsumer{
		consumer: consumer,
		handler:  handler,
	}

	consumer.RegisterHandler(TopicMatchesResults, mcc.handleMessage)

	return mcc
}

// handleMessage deserializes and routes a single Kafka message from matchmaking.matches.results.
func (mcc *MatchResultsCalculatedConsumer) handleMessage(ctx context.Context, msg *kafkago.Message) error {
	var event schemas.MatchmakingEvent
	if err := protojson.Unmarshal(msg.Value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal MatchmakingEvent (MatchResultsCalculated)",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"error", err)
		return nil
	}

	envelope := event.GetEnvelope()
	if envelope == nil {
		slog.ErrorContext(ctx, "MatchmakingEvent has nil envelope, skipping",
			"topic", msg.Topic,
			"offset", msg.Offset)
		return nil
	}

	data, ok := event.GetData().(*schemas.MatchmakingEvent_MatchResultsCalculated)
	if !ok || data.MatchResultsCalculated == nil {
		slog.WarnContext(ctx, "Unknown or missing MatchResultsCalculated payload, skipping",
			"event_id", envelope.GetId(),
			"event_type", envelope.GetType())
		return nil
	}

	if err := validateMatchResultsCalculatedResourceOwnership(envelope, data.MatchResultsCalculated); err != nil {
		slog.ErrorContext(ctx, "MatchResultsCalculated resource ownership validation failed, skipping",
			"event_id", envelope.GetId(),
			"match_id", data.MatchResultsCalculated.GetMatchId(),
			"error", err)
		return nil
	}

	slog.InfoContext(ctx, "Processing MatchResultsCalculated event",
		"event_id", envelope.GetId(),
		"match_id", data.MatchResultsCalculated.GetMatchId(),
		"resource_owner_id", envelope.GetResourceOwnerId())

	return mcc.handler(ctx, envelope, data.MatchResultsCalculated)
}

func validateMatchResultsCalculatedResourceOwnership(envelope *schemas.EventEnvelope, payload *schemas.MatchResultsCalculatedPayload) error {
	if strings.TrimSpace(payload.GetResourceOwnerId()) == "" {
		return fmt.Errorf("%w: resource_owner_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetMatchId()) == "" {
		return fmt.Errorf("%w: match_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetTenantId()) == "" {
		return fmt.Errorf("%w: tenant_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	if strings.TrimSpace(payload.GetClientId()) == "" {
		return fmt.Errorf("%w: client_id is empty in payload", ErrResourceOwnershipInvalid)
	}
	return nil
}

// Start begins consuming messages from matchmaking.matches.results.
func (mcc *MatchResultsCalculatedConsumer) Start(ctx context.Context) error {
	return mcc.consumer.Start(ctx)
}

// Close closes the consumer.
func (mcc *MatchResultsCalculatedConsumer) Close() error {
	return mcc.consumer.Close()
}

// PrizeDistributionConsumer consumes MatchResultsCalculated events from matchmaking.matches.results
// with group match-making-api-prize-distributor. Delegates to PrizeDistributionHandler for prize distribution.
// Uses the same topic and validation as MatchResultsCalculatedConsumer but separate consumer group.
type PrizeDistributionConsumer struct {
	inner *MatchResultsCalculatedConsumer
}

// NewPrizeDistributionConsumer creates a consumer for prize distribution (same topic, different group).
func NewPrizeDistributionConsumer(client *Client, groupID string, handler RatingsUpdatedHandler) *PrizeDistributionConsumer {
	return &PrizeDistributionConsumer{
		inner: NewMatchResultsCalculatedConsumer(client, groupID, handler),
	}
}

// Start begins consuming messages from matchmaking.matches.results.
func (p *PrizeDistributionConsumer) Start(ctx context.Context) error {
	return p.inner.Start(ctx)
}

// Close closes the consumer.
func (p *PrizeDistributionConsumer) Close() error {
	return p.inner.Close()
}

// AnalyticsTrackedConsumer consumes MatchResultsCalculated events from matchmaking.matches.results
// with group match-making-api-analytics-tracker. Produces AnalyticsTracked for analytics pipeline (#32).
type AnalyticsTrackedConsumer struct {
	inner *MatchResultsCalculatedConsumer
}

// NewAnalyticsTrackedConsumer creates a consumer for analytics (same topic, different group).
func NewAnalyticsTrackedConsumer(client *Client, groupID string, handler RatingsUpdatedHandler) *AnalyticsTrackedConsumer {
	return &AnalyticsTrackedConsumer{
		inner: NewMatchResultsCalculatedConsumer(client, groupID, handler),
	}
}

// Start begins consuming messages from matchmaking.matches.results.
func (a *AnalyticsTrackedConsumer) Start(ctx context.Context) error {
	return a.inner.Start(ctx)
}

// Close closes the consumer.
func (a *AnalyticsTrackedConsumer) Close() error {
	return a.inner.Close()
}
