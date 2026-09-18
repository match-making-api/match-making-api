# Game Configuration & Lobby Params

Refs **2508-005**. Source of truth for game rules and lobby creation parameters lives in **match-making-api** Mongo (`games`, `game_modes`). replay-api (and lobby/tournament flows) consume the resolved configuration via HTTP.

## Ownership

| Who | Responsibility |
|-----|----------------|
| Tenant/client admins via MM REST | Create/update `Game` and `GameMode` (resource_owner on BaseEntity) |
| match-making-api | Resolve merged config; expose lobby params |
| replay-api | Call configuration endpoint when creating queue/lobby; do not invent rules locally |

Updates without tenant+client on the mode are rejected by existing ownership validation on create (resource owner from context).

## Schema

### `Game` (rules / settings)

- `map_pool`, `allowed_regions`, `custom_rules`
- Team sizing: `min_players_per_team`, `max_players_per_team`, `number_of_teams`
- `skill_based_matching`, `enabled`

### `GameMode.lobby` (LobbyParams)

| Field | Meaning |
|-------|---------|
| `party_size` | Players per party (default 1) |
| `max_players` | Lobby capacity; default = game teams × max per team |
| `team_count` | Number of teams |
| `map_pool` | Optional override of game map pool |
| `allowed_regions` | Optional override of game regions |
| `custom_rules` | Merged over game rules (mode wins on key clash) |
| `ready_check_seconds` | Optional readiness window |

### Resolved `GameConfiguration`

Returned by `GET /game-modes/{id}/configuration`:

```json
{
  "game_id": "...",
  "game_mode_id": "...",
  "game_name": "...",
  "mode_name": "...",
  "tenant_id": "...",
  "client_id": "...",
  "enabled": true,
  "map_pool": ["..."],
  "regions": ["..."],
  "rules": {"key": "value"},
  "lobby": { "party_size": 1, "max_players": 2, "team_count": 2 }
}
```

## Exchange with replay-api

1. replay-api (or gateway) authenticates (JWT/RID).
2. Before queue join / lobby create, `GET /game-modes/{id}/configuration`.
3. Use `lobby.*`, `map_pool`, `regions`, `rules` when producing PlayerQueued or creating lobby.
4. Match-making match formation (2507+) should use the same resolver internally.

## Code

- Entities: `pkg/domain/game/entities/lobby_params.go`
- Resolver: `pkg/domain/game/usecases/resolve_game_configuration.go`
- HTTP: `GET /game-modes/{id}/configuration`
