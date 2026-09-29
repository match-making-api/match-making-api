# Tournament entry fee (2506-003)

When a player joins a **tournament** lobby whose prize pool `entry_fee_cents` is greater than zero, match-making publishes `ChargeableOperationRequested` on `matchmaking.billing.chargeable`. Wallet API consumes that event, validates the subscription, and collects the fee. Match-making does not call billing RPCs and does not update balances or usage.

## When published

| Lobby | Entry fee | Publish? |
|-------|-----------|----------|
| Not `tournament` | any | No |
| `tournament` | absent or `<= 0` | No |
| `tournament` | `> 0` | Yes, after the player is seated |

A player who is already in the lobby is rejected before a second event (`already_joined`).

## Payload

`operation_type` is `tournament_entry`.

| Field | Source |
|-------|--------|
| `amount_cents` | `prize_pool.entry_fee_cents` |
| `currency` | `USD` |
| `player_id` | authenticated user |
| `tournament_id` | lobby id |
| `resource_owner_id` | lobby resource owner user, or creator |
| `correlation_id` | request correlation when present |
| `idempotency_key` | `tournament_entry:<player_id>:<tournament_id>` |

## Payment failure

Match-making does not wait for wallet. If Kafka publish fails, the seat stays and the error is logged. Wallet must not charge a fee it never received. If wallet rejects the fee (upgrade required, limit exceeded, payment failed), that result stays in wallet. This story does not roll the player back out of the lobby. JoinTournament (2507-002) can add that compensation later.

## Code

- `BuildTournamentEntryFeeEvent` / `EntryFeeInputFromLobby`
- `LobbyController.Join` after a successful update
