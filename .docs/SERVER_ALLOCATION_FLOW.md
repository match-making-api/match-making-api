# Server Allocation and MatchReady Flow

Flow: MatchCreated → (replay-api allocates server) → ServerAllocated → match-making-api consumes → MatchReady broadcast to players (Epic §10 Phase 2).

## Flow

1. **MatchCreated** (match-making-api → replay-api): match exists with placeholder `game_server.server_id`.
2. **replay-api** (or game server manager) consumes MatchCreated, requests game server allocation.
3. **Game server service** allocates server, produces **ServerAllocated** to `matchmaking.server.allocated`.
4. **match-making-api** consumes ServerAllocated, validates resource ownership, produces **MatchReady** via `websocket.broadcasts`.
5. **replay-api** consumes `websocket.broadcasts`, delivers MatchReady to each player's WebSocket (TargetIDs).

## ServerAllocated Schema

- `match_id`, `server_id`, `region`, `resource_owner_id` (required)
- `player_ids` (required): players to receive MatchReady
- `connection_url`, `connection_token` (optional): connection details; avoid logging

## Resource Ownership Validation

| Rule | Description |
|------|-------------|
| resource_owner_id | Required in envelope and payload |
| match_id, server_id | Required |
| player_ids | Non-empty; only these players receive MatchReady |

Unauthorized events are skipped and logged.

## Failure Handling

| Failure | Handling |
|---------|----------|
| Validation failure | Log, skip (return nil), commit offset |
| Publish MatchReady failure | Return error, message not committed, retry |
| Malformed message | Log, skip, commit |
| Invalid player_id | Log, skip that player, continue |

**DLQ:** Consider routing repeated failures to `matchmaking.dlq` (future enhancement).

## Architecture

- **Producer (ServerAllocated):** game server service or replay-api. Must include `player_ids` from MatchCreated.
- **Consumer:** match-making-api (`consumer-server-allocated` binary).
- **MatchReady delivery:** `websocket.broadcasts` with `TargetIDs: [player_ids]`; replay-api routes to WebSocket connections.
