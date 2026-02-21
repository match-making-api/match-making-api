# Server Allocation Queue Flow (#queue-for-server)

When **MatchCreated** is produced and no game server is immediately available, matches are **enqueued** instead of failing. When a server becomes available, the game server manager pulls the next match and allocates it.

## Flow

1. **MatchCreated** (match-making-api → replay-api): match is created with placeholder `game_server.server_id`.
2. **Enqueue**: match-making-api enqueues the match in `ServerAllocationQueueStore` (Redis, FIFO per game_id:region).
3. **WaitingForServer**: broadcast to all players via `websocket.broadcasts` (Type: `WAITING_FOR_SERVER`).
4. **Game server manager**: when a server is free, calls `GET /server-allocation/next?game_id=X&region=Y`.
5. **Dequeue**: match-making-api returns the next match and removes it from the queue.
6. **Allocate**: game server manager allocates the server, produces **ServerAllocated**.
7. **MatchReady**: match-making-api consumes ServerAllocated, broadcasts MatchReady (#26).

## Ordering

- **FIFO** per `game_id` + `region`.
- Redis List: `matchmaking:server_allocation_queue:{game_id}:{region}`.

## Timeout & Abandon

- **Default timeout**: 10 minutes.
- **Worker**: `worker-server-allocation-timeout` runs every 60 seconds.
- **Abandon**: removes match from queue, broadcasts `SERVER_ALLOCATION_TIMEOUT` to players.
- **Config**: `ServerAllocationTimeoutWorkerConfig` (TimeoutMinutes, IntervalSeconds).

## Visibility

- **WaitingForServer**: when match is enqueued, players receive `WAITING_FOR_SERVER` via WebSocket.
- **ServerAllocationTimeout**: when match is abandoned, players receive `SERVER_ALLOCATION_TIMEOUT`.

## REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/server-allocation/next` | GET | Returns next match for allocation. Query: `game_id`, `region`. 200 with match or 204 if empty. |

## References

- Issue #queue-for-server
- `.docs/SERVER_ALLOCATION_FLOW.md` — ServerAllocated → MatchReady
- `.docs/MATCH_CREATION_FLOW.md` — MatchCreated production
