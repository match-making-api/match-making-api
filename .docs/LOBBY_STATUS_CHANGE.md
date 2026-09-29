# Lobby status change (2507-004)

Match-making publishes `LOBBY_STATUS_CHANGED` on the existing topic `matchmaking.lobby.events` when a lobby is created, a player joins, or the lobby is cancelled.

replay-api consumes the event and updates UI state (WebSocket to lobby participants). Match-making does not implement that consumer.

## Payload

| Field | Meaning |
|-------|---------|
| `lobby_id` | Lobby id, also the Kafka key |
| `event_type` | `LOBBY_STATUS_CHANGED` |
| `status` | `open`, `ready_check`, `starting`, `started`, `cancelled`, `completed` |
| `player_ids` | Seated players |
| `tenant_id`, `client_id`, `resource_owner_id` | Resource ownership |
| `metadata.reason` | `created`, `player_joined`, or `cancelled` |
| `game_type`, `region` | Lobby game and region |

A publish failure is logged. The HTTP write that changed the lobby still succeeds.

## When

| Action | Reason |
|--------|--------|
| `POST /api/lobbies` | `created` |
| `POST /api/lobbies/{id}/join` | `player_joined` |
| `DELETE /api/lobbies/{id}` | `cancelled` |
