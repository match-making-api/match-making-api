# Join tournament (2507-002)

`POST /api/lobbies/{id}/join` on a lobby with type `tournament` runs `JoinTournamentUseCase`.

## Order

1. Require caller tenant and client. They must match the lobby resource owner.
2. Reject closed, full, or already seated players with no Kafka publish.
3. If `entry_fee_cents` is greater than zero, publish `ChargeableOperationRequested` (`tournament_entry`) on `matchmaking.billing.chargeable`. A publish failure does not seat the player.
4. Save the seat.
5. Publish `PLAYER_JOINED_TOURNAMENT` on `matchmaking.lobby.events` with `lobby_id`, `player_ids`, and metadata `tenant_id`, `client_id`, `resource_owner_id`, `status`.

Match-making does not charge the fee. Wallet API consumes the chargeable event.

An already seated player returns `already_joined` and does not emit a second fee or a second `PLAYER_JOINED_TOURNAMENT`.
