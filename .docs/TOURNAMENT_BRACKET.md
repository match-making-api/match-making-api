# Tournament bracket and start match (2507-003)

`POST /api/lobbies/{id}/start-match` runs `StartTournamentMatchUseCase`.

## Bracket

The only bracket in this story is single elimination. Round 1 pairs seated players in slot order. A leftover player is a bye. The slate is stored on the lobby and is the audit record.

## Start

1. Caller tenant and client must match the lobby owner.
2. The lobby must be an open tournament with at least two seated players.
3. The first pending pair is marked `started` and receives a match id.
4. Prize pool status becomes `locked`. `PRIZE_POOL_UPDATED` is published on `matchmaking.prizepool.events`.
5. Lobby status becomes `starting`.
6. If `game_id` is a UUID and the server-allocation enqueuer is registered, the match is handed to match formation (2503) via `EnqueueAndBroadcast`.

Match-making does not publish `MatchStarted`. ADR-001 keeps that event inbound from the game server on `matchmaking.match.started`. Completion and prize transfer stay on the 2504 path after `MatchCompleted`.

A second start while the lobby is already `starting` returns `match already started` and does not publish another prize-pool event.
