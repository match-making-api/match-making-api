# MatchStarted Flow (#27)

End-to-end flow from MatchStarted event → validation → broadcast to match participants via WebSocket.

## Flow

1. **MatchStarted** (producer: game server / replay-api): Emitted when a match is about to begin (e.g. countdown finished, all players ready).
2. **Topic**: `matchmaking.match.started`
3. **Consumer**: `consumer-match-started` consumes, validates resource ownership and payload.
4. **Broadcast**: Publishes to `websocket.broadcasts` with `Type: MATCH_STARTED`, `TargetIDs: player_ids`, `LobbyID: match_id`.
5. **Delivery**: replay-api delivers the payload to each player's WebSocket connection.

## Payload (Proto)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `match_id` | string | Yes | Match identifier |
| `resource_owner_id` | string | Yes | RID for multi-tenancy/authorization |
| `player_ids` | string[] | Yes | Player IDs to notify (non-empty) |
| `countdown_seconds` | int32 | No | Optional countdown before start |
| `start_timestamp_epoch_ms` | int64 | No | Optional start timestamp |

## Validation Rules

| Rule | Description |
|------|-------------|
| ResourceOwnershipRule | `resource_owner_id` must be present in envelope and payload |
| MatchIDRule | `match_id` must be non-empty and valid UUID |
| PlayerIDsRule | `player_ids` must be non-empty; invalid UUIDs are skipped |

## Failure Modes

| Failure | Handling |
|---------|----------|
| Validation failure | Log error, skip processing. Message is not committed (consumer may retry). |
| No valid player_ids | Log error, return nil (don't retry — invalid data). |
| Invalid match_id | Log error, return nil (don't retry). |
| Publish failure | Return error to trigger consumer retry. |

## Idempotency

- MatchStarted keyed by `match_id` for Kafka partitioning.
- replay-api may dedupe by `match_id` + `event_id` when delivering to clients.

## References

- Issue #27 — MatchStarted flow
- `.docs/EVENT_SCHEMAS.md` — MatchStarted schema
- `pkg/infra/kafka/match_started_consumer.go` — Consumer implementation
- `pkg/domain/pairing/usecases/match_started_handler.go` — Handler implementation
