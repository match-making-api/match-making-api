package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	game_out "github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// AddAndFindNextPairExecutor defines the interface for adding and finding pairs
type AddAndFindNextPairExecutor interface {
	Execute(payload FindPairPayload) (*pairing_entities.Pair, *pairing_entities.Pool, int, error)
}

// EventPublisherInterface defines the interface for publishing events
type EventPublisherInterface interface {
	PublishMatchCreated(ctx context.Context, event *kafka.MatchEvent) error
}

// MatchmakingEventConsumer consumes events from replay-api and processes them
type MatchmakingEventConsumer struct {
	addAndFindNextPair AddAndFindNextPairExecutor
	eventPublisher     EventPublisherInterface
	regionReader       game_out.RegionReader
	poolReader         pairing_out.PoolReader
	poolWriter         pairing_out.PoolWriter
}

// NewMatchmakingEventConsumer creates a new consumer for matchmaking events
func NewMatchmakingEventConsumer(
	addAndFindNextPair AddAndFindNextPairExecutor,
	eventPublisher EventPublisherInterface,
	regionReader game_out.RegionReader,
	poolReader pairing_out.PoolReader,
	poolWriter pairing_out.PoolWriter,
) *MatchmakingEventConsumer {
	return &MatchmakingEventConsumer{
		addAndFindNextPair: addAndFindNextPair,
		eventPublisher:     eventPublisher,
		regionReader:       regionReader,
		poolReader:         poolReader,
		poolWriter:         poolWriter,
	}
}

// HandleQueueEvent processes queue join/leave events
func (c *MatchmakingEventConsumer) HandleQueueEvent(ctx context.Context, event *kafka.QueueEvent) error {
	slog.InfoContext(ctx, "Processing queue event",
		"event_type", event.EventType,
		"player_id", event.PlayerID,
		"game_type", event.GameType,
		"region", event.Region,
		"mmr", event.MMR)

	switch event.EventType {
	case kafka.EventTypeQueueJoined:
		return c.handleQueueJoined(ctx, event)
	case kafka.EventTypeQueueLeft:
		return c.handleQueueLeft(ctx, event)
	default:
		slog.WarnContext(ctx, "Unknown queue event type", "event_type", event.EventType)
		return nil
	}
}

// HandleLobbyEvent processes lobby events
func (c *MatchmakingEventConsumer) HandleLobbyEvent(ctx context.Context, event *kafka.LobbyEvent) error {
	slog.InfoContext(ctx, "Processing lobby event",
		"event_type", event.EventType,
		"lobby_id", event.LobbyID,
		"player_ids", event.PlayerIDs,
		"game_type", event.GameType,
		"region", event.Region)

	switch event.EventType {
	case kafka.EventTypePlayerJoined:
		slog.InfoContext(ctx, "Players joined lobby",
			"lobby_id", event.LobbyID,
			"player_ids", event.PlayerIDs,
			"game_type", event.GameType,
			"region", event.Region,
			"avg_mmr", event.AvgMMR)
	default:
		slog.WarnContext(ctx, "Unknown lobby event type", "event_type", event.EventType)
	}

	return nil
}

// handleQueueJoined processes a player joining the matchmaking queue
func (c *MatchmakingEventConsumer) handleQueueJoined(ctx context.Context, event *kafka.QueueEvent) error {
	slog.InfoContext(ctx, "Player joined matchmaking queue",
		"player_id", event.PlayerID,
		"game_type", event.GameType,
		"region", event.Region,
		"mmr", event.MMR)

	gameID, err := uuid.Parse(event.GameType)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid game type UUID", "game_type", event.GameType, "error", err)
		return err
	}

	// Lookup region by slug
	regions, err := c.regionReader.Search(ctx, map[string]interface{}{"slug": event.Region})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to lookup region", "region_slug", event.Region, "error", err)
		return err
	}
	if len(regions) == 0 {
		slog.ErrorContext(ctx, "Region not found", "region_slug", event.Region)
		return fmt.Errorf("region not found: %s", event.Region)
	}
	region := regions[0]
	
	payload := FindPairPayload{
		PartyID: event.PlayerID,
		Criteria: pairing_value_objects.Criteria{
			GameID: &gameID,
			Region: region,
			PairSize: 2, // Default to 1v1 for now
			SkillRange: &pairing_value_objects.SkillRange{
				MinMMR: event.MMR - 200,
				MaxMMR: event.MMR + 200,
			},
		},
	}

	pair, pool, position, err := c.addAndFindNextPair.Execute(payload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to add player to matchmaking pool", "error", err, "player_id", event.PlayerID)
		return err
	}

	if pair != nil {
		slog.InfoContext(ctx, "Match found!",
			"pair_id", pair.ID,
			"players", pair.Match,
			"player_id", event.PlayerID)

		// Extract player IDs from the pair
		playerIDs := make([]uuid.UUID, 0, len(pair.Match))
		for playerID := range pair.Match {
			playerIDs = append(playerIDs, playerID)
		}

		// Publish match created event
		matchEvent := &kafka.MatchEvent{
			MatchID:   pair.ID,
			LobbyID:   pair.ID, // Assuming lobby ID is the pair ID for now
			EventType: kafka.EventTypeMatchCreated,
			GameType:  event.GameType,
			Region:    event.Region,
			PlayerIDs: playerIDs,
		}
		if err := c.eventPublisher.PublishMatchCreated(ctx, matchEvent); err != nil {
			slog.ErrorContext(ctx, "Failed to publish match created event", "error", err, "pair_id", pair.ID)
			// Don't return error to avoid failing the matchmaking process
		}
	} else {
		slog.InfoContext(ctx, "Player added to pool",
			"pool_size", len(pool.Parties),
			"position", position,
			"player_id", event.PlayerID)
	}

	return nil
}

// handleQueueLeft processes a player leaving the matchmaking queue
func (c *MatchmakingEventConsumer) handleQueueLeft(ctx context.Context, event *kafka.QueueEvent) error {
	slog.InfoContext(ctx, "Player left matchmaking queue",
		"player_id", event.PlayerID,
		"game_type", event.GameType)

	gameID, err := uuid.Parse(event.GameType)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid game type UUID", "game_type", event.GameType, "error", err)
		return err
	}

	// Lookup region by slug
	regions, err := c.regionReader.Search(ctx, map[string]interface{}{"slug": event.Region})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to lookup region", "region_slug", event.Region, "error", err)
		return err
	}
	if len(regions) == 0 {
		slog.ErrorContext(ctx, "Region not found", "region_slug", event.Region)
		return fmt.Errorf("region not found: %s", event.Region)
	}
	region := regions[0]

	criteria := pairing_value_objects.Criteria{
		GameID:   &gameID,
		Region:   region,
		PairSize: 2, // Assuming same pair size
	}

	// Find the pool
	pool, err := c.poolReader.FindPool(&criteria)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find pool for removal", "error", err, "player_id", event.PlayerID)
		return err
	}
	if pool == nil {
		slog.WarnContext(ctx, "Pool not found for player removal", "player_id", event.PlayerID)
		return nil // Player not in any pool
	}

	// Remove the player from the pool
	_, err = pool.Remove(event.PlayerID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to remove player from pool", "error", err, "player_id", event.PlayerID)
		return err
	}

	// Save the updated pool
	_, err = c.poolWriter.Save(pool)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save pool after removal", "error", err, "player_id", event.PlayerID)
		return err
	}

	slog.InfoContext(ctx, "Player removed from matchmaking pool", "player_id", event.PlayerID)

	return nil
}

// HandlePlayerQueuedProto processes a PlayerQueued event from the canonical protobuf/CloudEvents schema (#16/#17).
// This is the handler for events consumed from the matchmaking.commands topic.
//
// Flow:
//  1. Validate resource ownership from envelope
//  2. Parse player_id and game_id as UUIDs
//  3. Lookup region by slug
//  4. Map to FindPairPayload with criteria from the event
//  5. Add player to pool via AddAndFindNextPair
//  6. If a pair is found, publish MatchCreated event
func (c *MatchmakingEventConsumer) HandlePlayerQueuedProto(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
	slog.InfoContext(ctx, "Processing PlayerQueued proto event",
		"event_id", envelope.GetId(),
		"player_id", payload.GetPlayerId(),
		"game_id", payload.GetGameId(),
		"region", payload.GetRegion(),
		"resource_owner_id", envelope.GetResourceOwnerId())

	// Validate resource ownership (defense-in-depth — consumer validates too)
	if strings.TrimSpace(envelope.GetResourceOwnerId()) == "" {
		slog.ErrorContext(ctx, "PlayerQueued event missing resource_owner_id, skipping",
			"event_id", envelope.GetId())
		return fmt.Errorf("resource_owner_id is required")
	}

	// Parse player_id
	playerID, err := uuid.Parse(payload.GetPlayerId())
	if err != nil {
		slog.ErrorContext(ctx, "Invalid player_id UUID", "player_id", payload.GetPlayerId(), "error", err)
		return fmt.Errorf("invalid player_id: %w", err)
	}

	// Parse game_id
	gameID, err := uuid.Parse(payload.GetGameId())
	if err != nil {
		slog.ErrorContext(ctx, "Invalid game_id UUID", "game_id", payload.GetGameId(), "error", err)
		return fmt.Errorf("invalid game_id: %w", err)
	}

	// Lookup region by slug
	regions, err := c.regionReader.Search(ctx, map[string]interface{}{"slug": payload.GetRegion()})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to lookup region", "region_slug", payload.GetRegion(), "error", err)
		return fmt.Errorf("failed to lookup region %s: %w", payload.GetRegion(), err)
	}
	if len(regions) == 0 {
		slog.ErrorContext(ctx, "Region not found", "region_slug", payload.GetRegion())
		return fmt.Errorf("region not found: %s", payload.GetRegion())
	}
	region := regions[0]

	// Parse tenant_id and client_id for criteria
	var tenantID, clientID *uuid.UUID
	if tid, err := uuid.Parse(payload.GetTenantId()); err == nil {
		tenantID = &tid
	}
	if cid, err := uuid.Parse(payload.GetClientId()); err == nil {
		clientID = &cid
	}

	// Build criteria from event payload
	criteria := pairing_value_objects.Criteria{
		GameID:   &gameID,
		Region:   region,
		PairSize: 2, // Default to 1v1; TODO: derive from game mode or event metadata
		TenantID: tenantID,
		ClientID: clientID,
	}

	// Map skill range if provided
	if sr := payload.GetSkillRange(); sr != nil {
		criteria.SkillRange = &pairing_value_objects.SkillRange{
			MinMMR: int(sr.GetMinMmr()),
			MaxMMR: int(sr.GetMaxMmr()),
		}
	}

	// Map priority boost
	if pb := payload.PriorityBoost; pb != nil {
		criteria.PriorityBoost = *pb > 0
	}

	findPairPayload := FindPairPayload{
		PartyID:  playerID,
		Criteria: criteria,
	}

	pair, pool, position, err := c.addAndFindNextPair.Execute(findPairPayload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to add player to matchmaking pool",
			"error", err,
			"player_id", playerID,
			"event_id", envelope.GetId())
		return err
	}

	if pair != nil {
		slog.InfoContext(ctx, "Match found from PlayerQueued event!",
			"pair_id", pair.ID,
			"players", pair.Match,
			"event_id", envelope.GetId())

		// Extract player IDs from the pair
		playerIDs := make([]uuid.UUID, 0, len(pair.Match))
		for pid := range pair.Match {
			playerIDs = append(playerIDs, pid)
		}

		// Publish match created event via legacy EventPublisher
		matchEvent := &kafka.MatchEvent{
			MatchID:   pair.ID,
			LobbyID:   pair.ID,
			EventType: kafka.EventTypeMatchCreated,
			GameType:  payload.GetGameId(),
			Region:    payload.GetRegion(),
			PlayerIDs: playerIDs,
		}
		if err := c.eventPublisher.PublishMatchCreated(ctx, matchEvent); err != nil {
			slog.ErrorContext(ctx, "Failed to publish match created event", "error", err, "pair_id", pair.ID)
			// Don't return error — the match was created, just the notification failed
		}
	} else {
		slog.InfoContext(ctx, "Player added to pool from PlayerQueued event",
			"pool_size", len(pool.Parties),
			"position", position,
			"player_id", playerID,
			"event_id", envelope.GetId())
	}

	return nil
}