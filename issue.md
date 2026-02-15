## Context

Players join the matchmaking queue via **replay-api** (e.g. POST `/matchmaking/queue/join`). The queue state is maintained in **match-making-api**, which consumes **PlayerQueued** events from Kafka. This story implements the **end-to-end join flow**: API → produce event → consumer validates and adds player to pool, so that match-making can form matches from a consistent, event-sourced pool.

## Description

Implement the **player queue join** flow with **event-driven** communication between replay-api and match-making-api. When a player joins:

1. **replay-api** accepts the request, validates the player and resource ownership, and **produces a PlayerQueued event** to Kafka.
2. **match-making-api** **consumes** the event, **validates resource ownership**, and adds the player to the appropriate pool (or updates state).

The event must carry `player_id`, `game_id`, `region`, and `resource_owner_id` (and tenant/client context as per schema). Consumers must reject or skip events that fail ownership validation.

## Background

- Epic §10 "Phase 1: Player Queue Entry" describes: `POST /matchmaking/queue/join` → produce `PlayerQueued` → match-making-api consumes, stores player in pool, optionally produces `PlayerQueueConfirmed` → API returns 200 with queue position/ETA.
- **Resource ownership** is mandatory: multi-tenant isolation and authorization depend on validating `tenant_id`, `client_id`, and `resource_owner_id` on both producer and consumer sides.
- The pool model in match-making-api is party-based (Pool with Parties); the join flow may map a single “player” from replay-api to a Party or equivalent abstraction — document the mapping.

## Scope

**In scope:**

- **API**: `POST /matchmaking/queue/join` (or equivalent) in replay-api. Request body captures at least game, region, and any preferences (e.g. skill range, priority boost). Auth identifies the player and resource owner.
- **Producer**: Publish **PlayerQueued** to Kafka with `player_id`, `game_id`, `region`, `resource_owner_id`, and tenant/client fields per schema.
- **Consumer** (match-making-api): Consume PlayerQueued, **validate resource ownership**, then add/update player in pool. Handle duplicates (e.g. re-join) according to product rules.
- **Response**: API returns success (e.g. 200) with queue position and/or ETA if available; document how position is derived (e.g. from match-making-api via sync call or subsequent event).

**Out of scope:**

- Queue **leave** (#23) and **status updates** (#22).
- Match formation, server allocation, or post-match flows (2503, 2504).
- WebSocket push for queue position (#22); initial join response can be sync only if needed.

## Acceptance Criteria

- [ ] **POST /matchmaking/queue/join** (or agreed path) is implemented and **produces a PlayerQueued event** to Kafka on success.
- [ ] Event **contains** at least: `player_id`, `game_id`, `region`, `resource_owner_id`, and tenant/client context as per #16 schema.
- [ ] **Consumer** in match-making-api **validates resource ownership** before processing; invalid events are logged and skipped (or dead-lettered per #35).
- [ ] Player is **added to the pool** (or state updated) so that match-making can consider them for matches.
- [ ] API response and error cases (e.g. invalid game, already in queue, Kafka produce failure) are **documented** and **consistent** (e.g. HTTP status codes, error payloads).

## Technical Notes

- Reuse producer implementation from #17 ; this story focuses on the **join** API, event payload construction, and **consumer** logic.
- If **PlayerQueueConfirmed** (or similar) is used, define who consumes it and how the API derives queue position/ETA — sync vs async.
- Consider idempotency: same player joining twice in quick succession; define and implement behaviour (e.g. update position, or reject).

## Dependencies

- #15  (Kafka + topics), #16  (schemas), #17  (producers in replay-api).
- match-making-api consumer setup for `matchmaking.commands` / `matchmaking.events` (as per topic design).
- Auth middleware and resource-owner resolution in replay-api.

## References

- Epic §9 — PlayerQueued event payload.
- Epic §10 — Phase 1: Player Queue Entry, Real-Time Event Flows.
