# Tournament lobby creation (2507-001)

Creates a tournament lobby and prize pool in match-making-api, then publishes events. Wallet-api executes any creation fee. Match-making does not call billing.

## HTTP

`POST /api/lobbies/tournament`

Auth: JWT or RID (same middleware as other lobby routes). Tenant and client come from the verified context.

```json
{
  "tournament_id": "uuid",
  "name": "Weekend Cup",
  "game_id": "cs2",
  "region": "br",
  "max_players": 8,
  "distribution_rule": "winner_takes_all",
  "amount_cents": 10000,
  "currency": "USD",
  "creation_fee_cents": 0
}
```

`creation_fee_cents` omitted or 0: no chargeable event. Greater than 0: publish `lobby_creation_fee`.

## Persistence

Lobby `type=tournament` with nested `prize_pool`:

| Field | Meaning |
|-------|---------|
| amount_cents | Pool amount |
| distribution_rule | e.g. winner_takes_all, top_3 |
| status | `open` on create |
| tournament_id | External tournament id |
| creation_fee_cents | Organizer fee (not the pool) |

Resource owner (tenant, client, user) is required on write.

## Events

| Event | Topic |
|-------|--------|
| `LOBBY_CREATED` | `matchmaking.lobby.events` |
| `PRIZE_POOL_CREATED` | `matchmaking.prizepool.events` |
| `ChargeableOperationRequested` (`lobby_creation_fee`) | `matchmaking.billing.chargeable` |

Idempotency key: `lobby_creation_fee:<lobby_id>`.

## Code

- `pkg/domain/lobbies/usecase/create_tournament_lobby.go`
- `POST /api/lobbies/tournament`
