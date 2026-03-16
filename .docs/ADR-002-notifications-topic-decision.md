# ADR-002: Notifications Topic — Decision Record

| Field       | Value |
|-------------|-------|
| **Status**  | Accepted |
| **Date**    | 2026-02-17 |
| **Deciders** | Match-Making-API team |
| **Related** | Epic §10 Kafka Topic Architecture, Issue #20 (this story), #22 (QueueStatusUpdated), #26 (MatchReady), #27 (MatchStarted) |

## Context

The epic §10 references a **`matchmaking.notifications`** topic for user-facing real-time events (queue_status, match_ready). The platform already has a **`websocket.broadcasts`** topic that serves as the delivery channel for WebSocket messages — consumed by **replay-api**'s WebSocket hub, which delivers events directly to player connections.

Currently, **QueueStatusUpdated** events (#22) are produced directly to `websocket.broadcasts` by the `QueueStatusTicker` worker. Future events like **MatchReady** (#26) and **MatchStarted** (#27) will also target specific players in real-time.

We need to decide whether to create a dedicated `matchmaking.notifications` topic or continue routing through existing topics.

## Decision

**We will NOT create a `matchmaking.notifications` topic.** Real-time notification events are produced directly to **`websocket.broadcasts`**, which is already consumed by the WebSocket hub for player delivery.

## Rationale

### Why `websocket.broadcasts` is sufficient

1. **Direct delivery path.** All current notification events (`QueueStatusUpdated`) and planned ones (`MatchReady`, `MatchStarted`) are WebSocket-bound. The `websocket.broadcasts` topic is already consumed by `replay-api`'s WebSocket hub, which delivers events to player connections. Adding an intermediate `matchmaking.notifications` topic would require a **consumer to re-route** events to `websocket.broadcasts`, adding latency without functional benefit.

2. **`WebSocketBroadcastEvent` already supports targeted delivery.** The existing payload struct includes `TargetIDs []uuid.UUID` for player-specific delivery and `LobbyID` for lobby-scoped broadcasts. This covers all notification use cases:
   - `QueueStatusUpdated` → `TargetIDs: [playerID]`
   - `MatchReady` → `TargetIDs: [player1, player2, ...]` or `LobbyID: lobbyID`
   - `MatchStarted` → `LobbyID: lobbyID`

3. **Operational simplicity.** One fewer topic to create, monitor, tune partitions/retention, and secure with ACLs. At our scale (2 services, <5 notification event types), the overhead is not justified.

4. **Separation already exists at the event type level.** Each event carries a `type` field (`QUEUE_STATUS_UPDATED`, `MATCH_READY`, etc.) that consumers use for routing. Domain-level separation is achieved through event types, not topic segregation.

### When to reconsider

Introduce `matchmaking.notifications` if any of the following become true:

- **Non-WebSocket notification channels** are needed (email, push notifications, SMS) — a dedicated topic allows multiple consumers with different delivery strategies
- **Rate limiting or backpressure** is needed specifically for notifications without affecting other `websocket.broadcasts` traffic
- **>3 independent services** consume notification events with different SLAs
- **Audit or compliance** requires a persistent log of all user notifications separate from WebSocket delivery

### Migration path (if needed later)

1. Create `matchmaking.notifications` topic with appropriate partitions/retention
2. Update producers (`QueueStatusTicker`, future MatchReady/MatchStarted producers) to publish to `matchmaking.notifications` instead of `websocket.broadcasts`
3. Create a lightweight consumer that reads from `matchmaking.notifications` and republishes to `websocket.broadcasts` (or have the WebSocket hub consume directly from `matchmaking.notifications`)
4. Add consumers for other channels (email service, push notification service)

## Consequences

### Positive

- No additional topic to manage (partitions, retention, ACLs, monitoring)
- Lower end-to-end latency (events go directly to WebSocket delivery)
- Simpler producer code (single publish call)
- Existing `QueueStatusTicker` and planned producers need no changes

### Negative

- All notification events share `websocket.broadcasts` with other broadcast events (lobby updates, etc.)
- If non-WebSocket channels are needed, producers must be updated to publish to a new topic
- No separate retention/partitioning for notification events vs. other broadcasts

### Neutral

- Event types provide logical separation within the shared topic
- ACLs for `websocket.broadcasts` are already configured in the `matchmaking-producer` KafkaUser

## Notification Event Routing (Current Architecture)

```
match-making-api                          replay-api
┌──────────────────────┐                  ┌──────────────────────┐
│ QueueStatusTicker    │                  │ WebSocket Hub        │
│   (worker)           │──publish──┐      │                      │
│                      │           │      │ WebSocketBroadcast   │
│ MatchReady producer  │──publish──┼─────►│   Consumer           │──► Player WS
│   (future #26)       │           │      │                      │
│                      │           │      │ Routes by event type │
│ MatchStarted producer│──publish──┘      │ + TargetIDs/LobbyID  │
│   (future #27)       │                  │                      │
└──────────────────────┘                  └──────────────────────┘
            │                                        │
            └─── websocket.broadcasts topic ─────────┘
```

## Alternatives Considered

### Create `matchmaking.notifications` + consumer to route to `websocket.broadcasts`

- **Pros**: Clean domain separation; ready for multi-channel delivery; separate retention/partitioning
- **Cons**: Extra hop adds ~5-15ms latency; extra consumer to deploy/monitor; no current need for multi-channel
- **Verdict**: Overkill for current architecture; easy to add later if needed

### Create `matchmaking.notifications` and have WebSocket hub consume it directly

- **Pros**: Domain topic with direct consumption; no intermediate routing
- **Cons**: WebSocket hub already consumes `websocket.broadcasts` and `matchmaking.lobby.events`; adding a third topic increases consumer complexity; breaks existing convention
- **Verdict**: Cleaner than option 1 but still adds unnecessary topic at current scale

### Use `matchmaking.events` (generic domain events topic)

- **Pros**: Single topic for all matchmaking domain events
- **Cons**: Mixes domain events (MatchCreated, RatingsUpdated) with user-facing notifications (QueueStatus, MatchReady); different retention and consumer needs; high-throughput events would drown low-latency notifications
- **Verdict**: Wrong abstraction level; domain events and notifications serve different purposes

## Related Documents

- [ADR-001: Schema Registry Decision](ADR-001-schema-registry-decision.md)
- [EVENT_SCHEMAS.md](EVENT_SCHEMAS.md) — schema catalog, topic mapping
- [KAFKA_SECURITY.md](KAFKA_SECURITY.md) — KafkaUser ACLs for `websocket.broadcasts`
