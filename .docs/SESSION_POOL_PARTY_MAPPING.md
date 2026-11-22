# Session ↔ Pool ↔ Party mapping — Refs 2508-003

## Overview

replay-api **Session** (one person) publishes `PlayerQueued` / `PlayerLeftQueue`.
match-making-api maps that into **Party** (team unit in the Pool) and pool shard.

| Concept | Owner | Notes |
|---------|--------|------|
| Session | replay-api | One human connection / player |
| Party | match-making-api | Pool membership key (`party_id`) |
| Pool | match-making-api | Shard by game + region (+ tenant/client) |

## Rules

1. **Solo:** if `party_id` absent/invalid → `party_id = player_id`.
2. **Group:** producer sends shared `party_id` on `PlayerQueuedPayload.party_id` (and optionally on leave) for all members.
3. **Ownership:** `tenant_id`, `client_id`, `resource_owner_id` required; must stay consistent on leave.
4. **Leave:** `PlayerLeftQueue` applies the same mapping, then removes the player from pool and active queue (idempotent).

## Implementation

- Helper: `pkg/domain/pairing/mapping` (`ResolvePartyID`, `MapQueueJoin`)
- Proto: `PlayerQueuedPayload.party_id` / `PlayerLeftQueuePayload.party_id` (additive, optional)
- Consumer: `HandlePlayerQueuedProto` / `HandlePlayerLeftQueueProto` use `MapQueueJoin`
- Active queue tracks `party_id` for status broadcasts

## Tests

```bash
go test ./pkg/domain/pairing/mapping/...
```
