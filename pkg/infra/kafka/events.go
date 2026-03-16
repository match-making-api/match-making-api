package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"google.golang.org/protobuf/encoding/protojson"
)

// Topic constants for matchmaking events
const (
	TopicQueueEvents        = "matchmaking.queue.events"
	TopicLobbyEvents        = "matchmaking.lobby.events"
	TopicPrizePoolEvents    = "matchmaking.prizepool.events"
	TopicMatchesCreated     = "matchmaking.matches.created"
	TopicMatchesResults     = "matchmaking.matches.results"
	TopicPlayerStatus       = "matchmaking.player-status"
	TopicWebSocketBroadcast = "websocket.broadcasts"
	TopicDLQ                = "matchmaking.dlq"
	TopicReadyCheck         = "matchmaking.ready-check"      // Ready check lifecycle events
	TopicNotificationDelivery    = "notifications.delivery"    // Notification delivery commands
	TopicNotificationDeliveryDLQ = "notifications.delivery.dlq" // Failed notification delivery
	TopicNotificationStatus      = "notifications.status"      // Delivery status updates

	// TopicMatchmakingCommands is the topic for canonical protobuf/CloudEvents commands
	// from replay-api (e.g. PlayerQueued). Consumed by match-making-api.
	TopicMatchmakingCommands = "matchmaking.commands"

	// TopicPlayerQueueConfirmed is the topic for PlayerQueueConfirmed events
	// (match-making-api → replay-api). Emitted after adding player to pool.
	// replay-api consumes to respond 200 OK with position/ETA to the client.
	TopicPlayerQueueConfirmed = "matchmaking.queue.confirmed"

	// TopicServerAllocated is the topic for ServerAllocated events
	// (game server / replay-api → match-making-api). Emitted when a game server
	// is allocated for a match. match-making-api consumes and broadcasts MatchReady.
	TopicServerAllocated = "matchmaking.server.allocated"

	// TopicMatchCompleted is the topic for MatchCompleted events
	// (game server / replay-api → match-making-api). match-making-api consumes, calculates results, produces MatchResultsCalculated.
	TopicMatchCompleted = "matchmaking.matches.completed"

	// TopicMatchStarted is the topic for MatchStarted events
	// (game server / replay-api → match-making-api). Emitted when match is about to begin.
	// match-making-api consumes and broadcasts to all match participants via WebSocket.
	TopicMatchStarted = "matchmaking.match.started"

	// TopicRatingsUpdated is the topic for RatingsUpdated events
	// (match-making-api → replay-api). Emitted after consuming MatchResultsCalculated.
	// replay-api consumes for leaderboards and skill-based matchmaking.
	TopicRatingsUpdated = "matchmaking.ratings.updated"

	// TopicPrizeDistributed is the topic for PrizeDistributed events
	// (match-making-api → wallet-api). Emitted after consuming MatchResultsCalculated when match has prize pool.
	// Wallet API consumes and executes prize transfers.
	TopicPrizeDistributed = "matchmaking.prizes.distributed"

	// TopicAnalyticsTracked is the topic for AnalyticsTracked events
	// (match-making-api → analytics service). Emitted after consuming MatchResultsCalculated (#32).
	// Analytics service consumes for dashboards, reporting.
	TopicAnalyticsTracked = "matchmaking.analytics.tracked"
)

// Event types
const (
	EventTypeQueueJoined        = "QUEUE_JOINED"
	EventTypeQueueLeft          = "QUEUE_LEFT"
	EventTypeSearching          = "SEARCHING"
	EventTypeLobbyCreated       = "LOBBY_CREATED"
	EventTypeLobbyUpdated       = "LOBBY_UPDATED"
	EventTypePlayerJoined       = "PLAYER_JOINED"
	EventTypePlayerLeft         = "PLAYER_LEFT"
	EventTypeReadyStatusChanged = "READY_STATUS_CHANGED"
	EventTypeLobbyReady         = "LOBBY_READY"
	EventTypeLobbyCancelled     = "LOBBY_CANCELLED"
	EventTypePrizePoolUpdated   = "PRIZE_POOL_UPDATED"
	EventTypeMatchCreated       = "MATCH_CREATED"
	EventTypeMatchStarted       = "MATCH_STARTED"
	EventTypeMatchCompleted     = "MATCH_COMPLETED"
	EventTypeMatchCancelled     = "MATCH_CANCELLED"

	// Ready check event types
	EventTypeReadyCheckStarted       = "READY_CHECK_STARTED"
	EventTypeReadinessConfirmed      = "READINESS_CONFIRMED"
	EventTypeReadinessDeclined       = "READINESS_DECLINED"
	EventTypeReadyCheckTimeout       = "READY_CHECK_TIMEOUT"
	EventTypeAllPlayersReady         = "ALL_PLAYERS_READY"
	EventTypeGameConnectionDelivered = "GAME_CONNECTION_DELIVERED"
	EventTypeQueueStatusUpdated      = "QUEUE_STATUS_UPDATED"
	EventTypeMatchReady              = "MATCH_READY"               // (#26): lobby is full, match ready to start
	EventTypeWaitingForServer        = "WAITING_FOR_SERVER"        // Match enqueued for server allocation
	EventTypeServerAllocationTimeout = "SERVER_ALLOCATION_TIMEOUT" // Match abandoned after timeout
)

// QueueStatusPayload is the payload for QUEUE_STATUS_UPDATED events.
// Broadcast via WebSocket to the player to show live queue position changes.
type QueueStatusPayload struct {
	PlayerID        uuid.UUID `json:"player_id"`
	GameID          uuid.UUID `json:"game_id"`
	Region          string    `json:"region"`
	Position        int       `json:"position"`
	EstimatedWaitMs int64     `json:"estimated_wait_ms"`
	TotalInQueue    int       `json:"total_in_queue"`
	EventType       string    `json:"event_type"`
	Timestamp       int64     `json:"timestamp"`
}

// EventPublisher publishes domain events to Kafka topics
type EventPublisher struct {
	client *Client
}

// NewEventPublisher creates a new EventPublisher
func NewEventPublisher(client *Client) *EventPublisher {
	return &EventPublisher{client: client}
}

// QueueEvent represents a matchmaking queue event
type QueueEvent struct {
	EventID   uuid.UUID         `json:"event_id"`
	PlayerID  uuid.UUID         `json:"player_id"`
	GameType  string            `json:"game_type"`
	Region    string            `json:"region"`
	MMR       int               `json:"mmr"`
	QueueTime int64             `json:"queue_time"`
	EventType string            `json:"event_type"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// PublishQueueEvent publishes a queue event
func (p *EventPublisher) PublishQueueEvent(ctx context.Context, event *QueueEvent) error {
	event.EventID = uuid.New()
	if event.QueueTime == 0 {
		event.QueueTime = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.PlayerID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"game_type":  event.GameType,
		},
	}

	return p.client.Publish(ctx, TopicQueueEvents, msg)
}

// LobbyEvent represents a lobby lifecycle event
type LobbyEvent struct {
	EventID   uuid.UUID         `json:"event_id"`
	LobbyID   uuid.UUID         `json:"lobby_id"`
	EventType string            `json:"event_type"`
	PlayerIDs []uuid.UUID       `json:"player_ids,omitempty"`
	GameType  string            `json:"game_type"`
	Region    string            `json:"region"`
	AvgMMR    int               `json:"avg_mmr"`
	CreatedAt int64             `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// PublishLobbyEvent publishes a lobby event
func (p *EventPublisher) PublishLobbyEvent(ctx context.Context, event *LobbyEvent) error {
	event.EventID = uuid.New()
	if event.CreatedAt == 0 {
		event.CreatedAt = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.LobbyID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"lobby_id":   event.LobbyID.String(),
		},
	}

	return p.client.Publish(ctx, TopicLobbyEvents, msg)
}

// PrizePoolEvent represents a prize pool update event
type PrizePoolEvent struct {
	EventID        uuid.UUID         `json:"event_id"`
	PoolID         uuid.UUID         `json:"pool_id"`
	LobbyID        uuid.UUID         `json:"lobby_id"`
	EventType      string            `json:"event_type"`
	TotalAmount    int64             `json:"total_amount"`
	Currency       string            `json:"currency"`
	ContributorID  *uuid.UUID        `json:"contributor_id,omitempty"`
	ContributionAmt int64            `json:"contribution_amount,omitempty"`
	Timestamp      int64             `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PublishPrizePoolEvent publishes a prize pool event
func (p *EventPublisher) PublishPrizePoolEvent(ctx context.Context, event *PrizePoolEvent) error {
	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.PoolID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"lobby_id":   event.LobbyID.String(),
		},
	}

	return p.client.Publish(ctx, TopicPrizePoolEvents, msg)
}

// MatchEvent represents a match creation or result event
type MatchEvent struct {
	EventID   uuid.UUID         `json:"event_id"`
	MatchID   uuid.UUID         `json:"match_id"`
	LobbyID   uuid.UUID         `json:"lobby_id"`
	EventType string            `json:"event_type"`
	GameType  string            `json:"game_type"`
	Region    string            `json:"region"`
	PlayerIDs []uuid.UUID       `json:"player_ids"`
	Teams     []TeamInfo        `json:"teams,omitempty"`
	Result    *MatchResult      `json:"result,omitempty"`
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// TeamInfo contains team details in a match
type TeamInfo struct {
	TeamID    uuid.UUID   `json:"team_id"`
	Name      string      `json:"name"`
	PlayerIDs []uuid.UUID `json:"player_ids"`
	Side      string      `json:"side,omitempty"` // e.g., "CT", "T" for CS2
}

// MatchResult contains match outcome details
type MatchResult struct {
	WinnerTeamID  *uuid.UUID        `json:"winner_team_id,omitempty"`
	IsDraw        bool              `json:"is_draw"`
	Scores        map[string]int    `json:"scores"` // team_id -> score
	Duration      int64             `json:"duration_seconds"`
	PlayerStats   []PlayerMatchStat `json:"player_stats,omitempty"`
	CompletedAt   int64             `json:"completed_at"`
}

// PlayerMatchStat contains individual player performance
type PlayerMatchStat struct {
	PlayerID uuid.UUID `json:"player_id"`
	Kills    int       `json:"kills"`
	Deaths   int       `json:"deaths"`
	Assists  int       `json:"assists"`
	Score    int       `json:"score"`
	MMRChange int      `json:"mmr_change"`
}

// PublishPlayerQueueConfirmed publishes a PlayerQueueConfirmed event (match-making-api → replay-api).
// Emitted after adding player to pool. replay-api consumes to respond 200 OK with position/ETA.
func (p *EventPublisher) PublishPlayerQueueConfirmed(ctx context.Context, event *schemas.MatchmakingEvent) error {
	value, err := protojson.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal PlayerQueueConfirmed: %w", err)
	}

	key := ""
	if env := event.GetEnvelope(); env != nil {
		key = env.GetSubject()
	}
	if key == "" {
		if payload := event.GetPlayerQueueConfirmed(); payload != nil {
			key = payload.GetPlayerId()
		}
	}
	if key == "" {
		key = uuid.New().String()
	}

	headers := map[string]string{
		"ce_type":   schemas.EventTypePlayerQueueConfirmed,
		"ce_source": "match-making-api",
	}

	return p.client.PublishBytes(ctx, TopicPlayerQueueConfirmed, key, value, headers)
}

// PublishMatchCreatedProto publishes a MatchCreated event with full Protobuf payload (Epic §9).
// Includes match_id, players[], game_server (with resource ownership), lobby_id, tenant_id, client_id.
// Partition key: match_id for per-match ordering.
func (p *EventPublisher) PublishMatchCreatedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	value, err := protojson.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal MatchCreated: %w", err)
	}

	key := ""
	if payload := event.GetMatchCreated(); payload != nil {
		key = payload.GetMatchId()
	}
	if key == "" {
		key = uuid.New().String()
	}

	headers := map[string]string{
		"ce_type":   schemas.EventTypeMatchCreated,
		"ce_source": "match-making-api",
	}

	return p.client.PublishBytes(ctx, TopicMatchesCreated, key, value, headers)
}

// PublishMatchResultsCalculatedProto publishes a MatchResultsCalculated event to matchmaking.matches.results.
// Partition key: match_id for per-match ordering. Consumed by replay-api for ratings (#30) and prizes (#31).
func (p *EventPublisher) PublishMatchResultsCalculatedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	value, err := protojson.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal MatchResultsCalculated: %w", err)
	}

	key := ""
	if payload := event.GetMatchResultsCalculated(); payload != nil {
		key = payload.GetMatchId()
	}
	if key == "" {
		key = uuid.New().String()
	}

	headers := map[string]string{
		"ce_type":   schemas.EventTypeMatchResultsCalculated,
		"ce_source": "match-making-api",
	}

	return p.client.PublishBytes(ctx, TopicMatchesResults, key, value, headers)
}

// PublishRatingsUpdatedProto publishes a RatingsUpdated event to matchmaking.ratings.updated.
// Partition key: match_id. Consumed by replay-api for leaderboards and skill-based matchmaking.
func (p *EventPublisher) PublishRatingsUpdatedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	value, err := protojson.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal RatingsUpdated: %w", err)
	}

	key := ""
	if payload := event.GetRatingsUpdated(); payload != nil {
		key = payload.GetMatchId()
	}
	if key == "" {
		key = uuid.New().String()
	}

	headers := map[string]string{
		"ce_type":   schemas.EventTypeRatingsUpdated,
		"ce_source": "match-making-api",
	}

	return p.client.PublishBytes(ctx, TopicRatingsUpdated, key, value, headers)
}

// PublishPrizeDistributedProto publishes a PrizeDistributed event to matchmaking.prizes.distributed.
// Partition key: match_id. Consumed by wallet API for prize transfers.
func (p *EventPublisher) PublishPrizeDistributedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	value, err := protojson.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal PrizeDistributed: %w", err)
	}

	key := ""
	if payload := event.GetPrizeDistributed(); payload != nil {
		key = payload.GetMatchId()
	}
	if key == "" {
		key = uuid.New().String()
	}

	headers := map[string]string{
		"ce_type":   schemas.EventTypePrizeDistributed,
		"ce_source": "match-making-api",
	}

	return p.client.PublishBytes(ctx, TopicPrizeDistributed, key, value, headers)
}

// PublishAnalyticsTrackedProto publishes an AnalyticsTracked event to matchmaking.analytics.tracked.
// Partition key: match_id. Consumed by analytics service for dashboards and reporting (#32).
func (p *EventPublisher) PublishAnalyticsTrackedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	value, err := protojson.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal AnalyticsTracked: %w", err)
	}

	key := ""
	if payload := event.GetAnalyticsTracked(); payload != nil {
		key = payload.GetMatchId()
	}
	if key == "" {
		key = uuid.New().String()
	}

	headers := map[string]string{
		"ce_type":   schemas.EventTypeAnalyticsTracked,
		"ce_source": "match-making-api",
	}

	return p.client.PublishBytes(ctx, TopicAnalyticsTracked, key, value, headers)
}

// PublishMatchCreated publishes a match creation event (legacy JSON format)
func (p *EventPublisher) PublishMatchCreated(ctx context.Context, event *MatchEvent) error {
	event.EventID = uuid.New()
	event.EventType = EventTypeMatchCreated
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"lobby_id":   event.LobbyID.String(),
		},
	}

	return p.client.Publish(ctx, TopicMatchesCreated, msg)
}

// PublishMatchResult publishes a match result event
func (p *EventPublisher) PublishMatchResult(ctx context.Context, event *MatchEvent) error {
	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	msg := &Message{
		Key:       event.MatchID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"match_id":   event.MatchID.String(),
		},
	}

	return p.client.Publish(ctx, TopicMatchesResults, msg)
}

// WebSocketBroadcastEvent represents an event to broadcast to WebSocket clients
type WebSocketBroadcastEvent struct {
	EventID   uuid.UUID   `json:"event_id"`
	Type      string      `json:"type"`
	LobbyID   *uuid.UUID  `json:"lobby_id,omitempty"`
	TargetIDs []uuid.UUID `json:"target_ids,omitempty"` // specific user IDs, nil for broadcast
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

// PublishWebSocketBroadcast publishes an event for WebSocket broadcast
func (p *EventPublisher) PublishWebSocketBroadcast(ctx context.Context, event *WebSocketBroadcastEvent) error {
	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	key := "broadcast"
	if event.LobbyID != nil {
		key = event.LobbyID.String()
	}

	msg := &Message{
		Key:       key,
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.Type,
		},
	}

	return p.client.Publish(ctx, TopicWebSocketBroadcast, msg)
}

// PublishPlayerStatus publishes player status update (compacted topic)
func (p *EventPublisher) PublishPlayerStatus(ctx context.Context, playerID uuid.UUID, status string, metadata map[string]string) error {
	event := map[string]interface{}{
		"player_id":  playerID,
		"status":     status,
		"updated_at": time.Now().UnixMilli(),
		"metadata":   metadata,
	}

	msg := &Message{
		Key:       playerID.String(),
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"status": status,
		},
	}

	return p.client.Publish(ctx, TopicPlayerStatus, msg)
}

// PublishToDLQ publishes a failed message to the dead letter queue
func (p *EventPublisher) PublishToDLQ(ctx context.Context, originalTopic string, originalKey string, value interface{}, err error) error {
	dlqEvent := map[string]interface{}{
		"original_topic": originalTopic,
		"original_key":   originalKey,
		"value":          value,
		"error":          err.Error(),
		"timestamp":      time.Now().UnixMilli(),
	}

	msg := &Message{
		Key:       uuid.New().String(),
		Value:     dlqEvent,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"original_topic": originalTopic,
			"error_type":     "processing_failed",
		},
	}

	return p.client.Publish(ctx, TopicDLQ, msg)
}

// ReadyCheckEvent represents a readiness confirmation lifecycle event
type ReadyCheckEvent struct {
	EventID            uuid.UUID                                          `json:"event_id"`
	LobbyID            uuid.UUID                                          `json:"lobby_id"`
	PlayerID           *uuid.UUID                                         `json:"player_id,omitempty"`
	EventType          string                                             `json:"event_type"`
	PlayerIDs          []uuid.UUID                                        `json:"player_ids,omitempty"`
	Summary            interface{}                                        `json:"summary,omitempty"`
	GameConnectionInfo interface{}                                        `json:"game_connection_info,omitempty"`
	Timestamp          int64                                              `json:"timestamp"`
	Metadata           map[string]string                                  `json:"metadata,omitempty"`
}

// PublishReadyCheckEvent publishes a ready check lifecycle event
func (p *EventPublisher) PublishReadyCheckEvent(ctx context.Context, event *ReadyCheckEvent) error {
	event.EventID = uuid.New()
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	key := event.LobbyID.String()
	if event.PlayerID != nil {
		key = event.PlayerID.String()
	}

	msg := &Message{
		Key:       key,
		Value:     event,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": event.EventType,
			"lobby_id":   event.LobbyID.String(),
		},
	}

	return p.client.Publish(ctx, TopicReadyCheck, msg)
}

// PublishNotificationDelivery publishes a notification delivery command
func (p *EventPublisher) PublishNotificationDelivery(ctx context.Context, notificationID uuid.UUID, payload interface{}) error {
	msg := &Message{
		Key:       notificationID.String(),
		Value:     payload,
		Timestamp: time.Now(),
		Headers: map[string]string{
			"event_type": "NOTIFICATION_DELIVERY",
		},
	}

	return p.client.Publish(ctx, TopicNotificationDelivery, msg)
}
