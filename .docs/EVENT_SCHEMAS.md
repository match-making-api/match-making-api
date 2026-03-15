# Event Schemas — Matchmaking

Definition of canonical schemas for matchmaking events exchanged via Kafka between **replay-api** and **match-making-api**. The schemas ensure agreement on payload structure, field types, and resource ownership metadata.

> **Schema Registry**: We do **not** use a Schema Registry. Schemas are managed in-repo with Protobuf and versioned via `dataschema_version` in the envelope. See [ADR-001](ADR-001-schema-registry-decision.md) for the full decision record.

## Schema Location

- **Proto files**: `pkg/infra/events/schemas/matchmaking_events.proto`
- **Generated Go code**: `pkg/infra/events/schemas/matchmaking_events.pb.go`
- **Documentation**: `.docs/EVENT_SCHEMAS.md` (this file)
- **Decision record**: `.docs/ADR-001-schema-registry-decision.md`

## CloudEvents 1.0 Alignment

Events follow **CloudEvents 1.0** (https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md). The `EventEnvelope` maps to CloudEvents context attributes.

### Wire Format vs Proto Structure

- **Proto**: `MatchmakingEvent` uses `envelope` + `oneof data` (typed payloads). Serialization via `protojson` produces a nested JSON: `{"envelope": {...}, "playerQueued": {...}}`.
- **CloudEvents JSON format**: The spec defines a *flat* structure with top-level attributes and a single `data` member. Our nested structure is CloudEvents-inspired; for strict JSON compliance, consider flattening at serialization time or using `application/cloudevents+json` with a custom mapper.

### Context Attributes (CloudEvents 1.0)

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Unique event identifier (e.g. UUID). `source`+`id` must be unique per distinct event. |
| `type` | string | Yes | Event type. **Recommended**: reverse-DNS prefix (e.g. `com.leetgaming.matchmaking.PlayerQueued`). |
| `source` | URI-reference | Yes | Context where the event occurred. **Required**: non-empty URI-reference (e.g. `https://replay-api.example.com`, `/replay-api`). |
| `specversion` | string | Yes | CloudEvents spec version, e.g. `"1.0"`. |
| `time` | Timestamp (RFC 3339) | No | When the occurrence happened. |
| `subject` | string | No | Subject in producer context (e.g. `player_id`, `match_id`). |
| `datacontenttype` | string (RFC 2046) | No | Content type of `data` (e.g. `application/json`). Implicit when omitted in JSON. |
| `dataschema` | URI | No | Schema URI for `data`. We use extension `dataschema_version` instead. |

### Extension Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `resource_owner_id` | string | Yes | RID for multi-tenancy/authorization. |
| `correlation_id` | string | No | For distributed tracing. |
| `dataschema_version` | int32 | Yes | Payload schema version for evolution (distinct from CloudEvents `dataschema` URI). |

### Kafka Binding

When publishing to Kafka, producers may add CloudEvents headers per the Kafka protocol binding:

- `ce_type` — event type (maps to `type`)
- `ce_source` — event source (maps to `source`)

These enable routing and filtering without deserializing the message body.

### Recommendations for Full Compliance

1. **`source`**: Use a URI-reference (e.g. `https://replay-api.leet-gaming.com` or `/replay-api`), not a plain service name.
2. **`type`**: Use reverse-DNS prefix (e.g. `com.leetgaming.matchmaking.PlayerQueued`) for routing and namespacing.
3. **`datacontenttype`**: Set to `application/json` when payload is JSON (optional; implied when omitted).

## Defined Events

### PlayerQueued (replay-api → match-making-api)

Emitted when a player joins the queue.

**Payload:**
- `player_id`, `game_id`, `region` (required)
- `skill_range` (optional): min/max MMR
- `priority_boost` (optional): queue priority
- `tenant_id`, `client_id` (required)
- `resource_permissions` (optional): e.g. `["read", "write"]`

### PlayerLeftQueue (replay-api → match-making-api)

Emitted when a player leaves (cancels) the queue. Inverse of `PlayerQueued`. Consumer handles this idempotently — leaving when not in queue or duplicate leave events result in a no-op.

**Payload:**
- `player_id`, `game_id`, `region` (required)
- `tenant_id`, `client_id` (required)
- `reason` (optional): leave reason, e.g. `"user_cancelled"`, `"timeout"`, `"disconnect"`

**Consumer behavior:**
- Validates resource ownership (envelope + payload)
- Removes player from matchmaking pool (if present)
- Removes player from active queue store (Redis/Dragonfly)
- If player is not in pool or active queue, the operation is a no-op (idempotent)

### PlayerQueueConfirmed (match-making-api → replay-api)

Emitted after adding a player to the matchmaking pool. Enables the async round-trip: join → PlayerQueued → pool add → PlayerQueueConfirmed → 200 OK (position/ETA) to the client.

**Payload:**
- `player_id`, `game_id`, `region` (required)
- `position` (int32): queue position (1-based). 0 if match found immediately.
- `eta_seconds` (int32): best-effort estimated wait in seconds. 0 if match found.
- `tenant_id`, `client_id`, `resource_owner_id` (required)

**Consumer (replay-api) behavior:**
- Consume from `matchmaking.queue.confirmed`
- Use `correlation_id` from envelope to associate with the original join request
- Respond 200 OK with `{ position, eta_seconds }` to the client (or deliver via WebSocket/polling)

**Approach:** Async. The replay-api produces PlayerQueued and returns 202 Accepted (or holds the connection). When it consumes PlayerQueueConfirmed, it responds 200 OK with position/ETA. Alternative: sync via request-reply over Kafka (replay-api blocks until PlayerQueueConfirmed with matching correlation_id) — not implemented; async is preferred for scalability.

### MatchCreated (match-making-api → replay-api)

Emitted when a match is created. Flow: PotentialMatchFound (AddAndFindNextPair returns pair) → validation → MatchCreated.

**Payload:**
- `match_id`, `lobby_id`, `tenant_id`, `client_id` (required)
- `players[]`: `player_id`, `party_id`, `resource_permissions`
- `game_server`: `server_id` (placeholder until 2503-002), `region`, `resource_owner_id`

**Validation (before publish):**
- `resource_owner_id` present (envelope)
- `tenant_id`, `client_id` present (required for downstream access control)
- Pair exists with players and match_id

**Resource ownership transfer:** `game_server.resource_owner_id` and envelope carry ownership context for replay-api / game server access control.

**Failure modes:** Validation failure → log, skip MatchCreated (pair remains in DB). Produce failure → log, non-fatal (reconciliation may be needed).

### MatchCompleted (game server / replay-api → match-making-api)

Emitted when a match ends. match-making-api consumes from `matchmaking.matches.completed`, validates resource ownership, calculates results, persists, and produces MatchResultsCalculated.

**Payload:** `match_id`, `player_ids`, `winner_team_id`, `is_draw`, `completed_at_epoch_ms`, `tenant_id`, `client_id`.

**Consumer behavior:**
- Validates `resource_owner_id` in envelope; `match_id`, `tenant_id`, `client_id` in payload
- Calculates results (pass-through from MatchCompleted)
- Persists to MongoDB with resource ownership (idempotent by `match_id`)
- Produces MatchResultsCalculated to `matchmaking.matches.results`

### MatchResultsCalculated (match-making-api → replay-api)

Emitted after consuming MatchCompleted. Includes statistics for ratings (#30) and prizes (#31).

**Payload:** `match_id`, `player_ids`, `winner_team_id`, `is_draw`, `completed_at_epoch_ms`, `calculated_at_epoch_ms`, `tenant_id`, `client_id`, `resource_owner_id`.

**Topic:** `matchmaking.matches.results`

### RatingsUpdated (match-making-api → replay-api)

Emitted after consuming MatchResultsCalculated. Includes MMR deltas per player for leaderboards and skill-based matchmaking.

**Payload:** `match_id`, `deltas` (PlayerRatingDelta: player_id, mmr_before, mmr_after, delta), `updated_at_epoch_ms`, `tenant_id`, `client_id`, `resource_owner_id`, `game_id`, `audit_trail` (algorithm_version, reason, updated_by).

**Topic:** `matchmaking.ratings.updated`

**Consumer (replay-api) behavior:** Update leaderboards, persist ratings; validate resource ownership before applying.

### QueueStatusUpdated (match-making-api → replay-api, via `websocket.broadcasts`)

Emitted every 5s by the `worker-queue-status` for each active queue entry. Delivered to the player's WebSocket connection via `replay-api`.

**Payload** (`QueueStatusPayload` — JSON, not Protobuf):
- `player_id`, `game_id`, `region` (required)
- `position` (int): current queue position
- `estimated_wait_ms` (int64): estimated remaining wait in milliseconds
- `total_in_queue` (int): total players in the queue
- `event_type`: `"QUEUE_STATUS_UPDATED"`
- `timestamp` (int64): Unix epoch ms

### ServerAllocated (game server / replay-api → match-making-api)

Emitted when a game server is allocated for a match. match-making-api consumes, validates, broadcasts MatchReady.

**Payload:** `match_id`, `server_id`, `region`, `resource_owner_id`, `player_ids[]`, optional `connection_url`, `connection_token`.

### MatchReady (#26)

Emitted after ServerAllocated is consumed. Broadcast to all players in the match via `websocket.broadcasts` with `TargetIDs: [player_ids]`. Payload: `match_id`, `server_id`, `region`, `connection_url`, `connection_token`, `resource_owner_id`.

### MatchStarted (#27)

Emitted when a match is about to begin (e.g. countdown finished, all players ready). match-making-api consumes from `matchmaking.match.started`, validates, and broadcasts to all match participants via `websocket.broadcasts`.

**Payload (Proto):** `match_id`, `resource_owner_id`, `player_ids[]`, optional `countdown_seconds`, optional `start_timestamp_epoch_ms`.

**Broadcast:** `Type: MATCH_STARTED`, `TargetIDs: player_ids`, `LobbyID: match_id`.

## Topic → Schema Mapping

### Domain Events (Protobuf/CloudEvents)

| Topic | Direction | Events | Proto Message | `dataschema_version` |
|-------|-----------|--------|---------------|----------------------|
| `matchmaking.commands` | replay-api → match-making-api | PlayerQueued | `PlayerQueuedPayload` | 1 |
| `matchmaking.commands` | replay-api → match-making-api | PlayerLeftQueue | `PlayerLeftQueuePayload` | 1 |
| `matchmaking.queue.confirmed` | match-making-api → replay-api | PlayerQueueConfirmed | `PlayerQueueConfirmedPayload` | 1 |
| `matchmaking.matches.created` | match-making-api → replay-api | MatchCreated | `MatchCreatedPayload` | 1 |
| `matchmaking.server.allocated` | game server / replay-api → match-making-api | ServerAllocated | `ServerAllocatedPayload` | 1 |
| `matchmaking.match.started` | game server / replay-api → match-making-api | MatchStarted | `MatchStartedPayload` | 1 |
| `matchmaking.matches.completed` | game server / replay-api → match-making-api | MatchCompleted | `MatchCompletedPayload` | 1 |
| `matchmaking.matches.results` | match-making-api → replay-api | MatchResultsCalculated | `MatchResultsCalculatedPayload` | 1 |
| `matchmaking.ratings.updated` | match-making-api → replay-api | RatingsUpdated | `RatingsUpdatedPayload` | 1 |

### Real-Time Notifications (JSON, via `websocket.broadcasts`)

> **Decision**: We do NOT use a dedicated `matchmaking.notifications` topic. Notification events are published directly to `websocket.broadcasts` for WebSocket delivery. See [ADR-002](ADR-002-notifications-topic-decision.md).

| Topic | Direction | Events | Payload Format | Delivery |
|-------|-----------|--------|----------------|----------|
| `websocket.broadcasts` | match-making-api → replay-api | QueueStatusUpdated | `QueueStatusPayload` (JSON) | `TargetIDs: [playerID]` |
| `websocket.broadcasts` | match-making-api → replay-api | MatchReady (#26) | TBD | `TargetIDs: [playerIDs...]` or `LobbyID` |
| `websocket.broadcasts` | match-making-api → replay-api | MatchStarted (#27) | `MatchStartedBroadcastPayload` (JSON) | `TargetIDs: [playerIDs...]`, `LobbyID` |
| `websocket.broadcasts` | match-making-api → replay-api | LobbyUpdated, PlayerJoined, etc. | `WebSocketBroadcastEvent` (JSON) | `LobbyID` |

### Other Topics

| Topic | Direction | Events | Notes |
|-------|-----------|--------|-------|
| `matchmaking.queue.events` | match-making-api internal | QueueJoined, QueueLeft, Searching | Legacy queue events |
| `matchmaking.lobby.events` | match-making-api → replay-api | LobbyCreated, LobbyUpdated, etc. | Lobby lifecycle events |
| `matchmaking.player-status` | match-making-api → replay-api | Player status changes | Player online/offline |
| `matchmaking.dlq` | match-making-api internal | Failed events | Dead letter queue |

---

## Schema Versioning Strategy

### Overview

We use **repo-based Protobuf schemas** with an explicit `dataschema_version` field in the `EventEnvelope`. There is no external Schema Registry — see [ADR-001](ADR-001-schema-registry-decision.md).

### `dataschema_version` field

- Lives in `EventEnvelope.dataschema_version` (int32).
- **Initial version**: `1` for all current events.
- Producers **must** set this field on every event.
- Consumers **should** inspect this field and handle multiple versions when possible.

### Version lifecycle

| Version state | Description |
|--------------|-------------|
| **Current** | Actively produced and consumed. |
| **Deprecated** | Still consumed (backward compat) but new producers should use the next version. Announce deprecation in PR + this doc. |
| **Retired** | No longer consumed. Consumers may reject or skip. Remove support after all producers have migrated. |

### When to bump the version

| Change type | Bump? | Example |
|------------|-------|---------|
| Add optional field | No | Add `optional string nickname = 9` to `PlayerQueuedPayload` |
| Add new event type to `oneof` | No | Add `QueueStatusPayload queue_status = 14` to `MatchmakingEvent` |
| Change field type | **Yes** | Change `string player_id` to `bytes player_id` |
| Remove a field | **Yes** | Remove `priority_boost` (use `reserved` instead) |
| Rename with different `json_name` | **Yes** | Rename field AND change its `json_name` |
| Change field semantics | **Yes** | `region` changes from ISO country code to cloud region ID |

## Compatibility Rules

### Default policy: **Backward Compatible**

New schema versions **must** be readable by consumers running the **previous** version. This means:

- **Allowed**: Add new optional fields, add new `oneof` variants, add new enum values
- **Forbidden**: Remove fields, change field numbers, change field types, change required→optional

### Compatibility matrix

| Producer version | Consumer version | Result |
|-----------------|-----------------|--------|
| v1 | v1 | OK |
| v2 | v1 | OK — consumer ignores unknown fields (Protobuf default) |
| v1 | v2 | OK — consumer handles missing optional fields with defaults |
| v2 | v1 (field removed in v2) | **BREAK** — consumer expects field that no longer exists |

### How Protobuf guarantees backward compatibility

1. **Unknown fields are preserved.** A v1 consumer receiving a v2 message ignores fields it doesn't know about.
2. **Missing optional fields have default values.** A v2 consumer receiving a v1 message gets zero-values for new fields.
3. **Field numbers are the contract.** Renaming a field name is safe; the wire format uses numbers.

### Forward compatibility (optional, recommended)

For critical events (e.g. `PlayerQueued`), we also aim for **forward compatibility**: old producers can send messages that new consumers understand. This is naturally achieved by Protobuf's additive model.

## Schema Evolution Process

### Adding a new field (non-breaking)

1. **Add** the field to the proto message with a new, unused field number:

```protobuf
message PlayerQueuedPayload {
  // ... existing fields ...
  optional string platform = 9;  // NEW: player's platform (e.g. "steam", "xbox")
}
```

2. **Do NOT** bump `dataschema_version` (additive change).
3. Run `make proto-gen` to regenerate Go code.
4. Update producer to populate the new field.
5. Update consumer to read the new field (with a fallback for old messages that don't have it).
6. Update this document (topic → schema table, event description).
7. Open PR — reviewers verify backward compatibility.

### Adding a new event type (non-breaking)

1. **Define** the new payload message in `matchmaking_events.proto`.
2. **Add** it to the `oneof data` in `MatchmakingEvent` with a new number:

```protobuf
message MatchmakingEvent {
  EventEnvelope envelope = 1;
  oneof data {
    PlayerQueuedPayload player_queued = 10;
    MatchCreatedPayload match_created = 11;
    MatchCompletedPayload match_completed = 12;
    RatingsUpdatedPayload ratings_updated = 13;
    QueueStatusPayload queue_status = 14;  // NEW
  }
}
```

3. Run `make proto-gen`.
4. Implement producer and consumer logic.
5. Add topic → schema mapping to this document.
6. Open PR.

### Making a breaking change (version bump required)

Breaking changes should be **avoided** whenever possible. If unavoidable:

1. **Bump** `dataschema_version` (e.g. `1` → `2`).
2. **Document** the breaking change in this file under the affected event.
3. **Implement dual-version support** in the consumer:

```go
switch envelope.GetDataschemaVersion() {
case 1:
    // handle v1 payload
case 2:
    // handle v2 payload
default:
    slog.Warn("Unknown schema version", "version", envelope.GetDataschemaVersion())
}
```

4. **Coordinate rollout**:
   - Deploy consumer with dual-version support first.
   - Then deploy producer with new version.
   - After all producers are on v2, deprecate v1 in the consumer.
5. **Mark deprecated fields** with `reserved`:

```protobuf
message PlayerQueuedPayload {
  reserved 5;  // was: priority_boost (removed in v2)
  reserved "priority_boost";
}
```

6. Open PR — require explicit approval from both producer and consumer team leads.

### Deprecating a field (non-breaking, recommended approach)

Instead of removing a field, deprecate it:

1. Add a comment marking it deprecated:

```protobuf
message PlayerQueuedPayload {
  // ...
  optional int32 priority_boost = 5 [deprecated = true];  // Deprecated in v1.1 — use queue_priority instead
  optional int32 queue_priority = 9;  // Replaces priority_boost
}
```

2. Stop populating the deprecated field in producers.
3. Update consumers to prefer the new field, falling back to deprecated.
4. After a migration period, mark the field number as `reserved`.

## Schema Ownership

### Ownership model

| Schema | Owner | Can modify | Must approve changes |
|--------|-------|-----------|---------------------|
| `EventEnvelope` | Platform team | Platform team | Platform team lead |
| `PlayerQueuedPayload` | replay-api team | replay-api team | replay-api + match-making-api leads |
| `PlayerLeftQueuePayload` | replay-api team | replay-api team | replay-api + match-making-api leads |
| `MatchCreatedPayload` | match-making-api team | match-making-api team | match-making-api + replay-api leads |
| `MatchCompletedPayload` | match-making-api team | match-making-api team | match-making-api + replay-api leads |
| `RatingsUpdatedPayload` | match-making-api team | match-making-api team | match-making-api + replay-api leads |

### Rules

1. **The producer owns the schema.** The team that produces the event is responsible for its schema definition.
2. **Both sides approve changes.** Any change to a shared schema requires approval from both the producer and consumer team leads (enforced via GitHub CODEOWNERS or PR review rules).
3. **The `EventEnvelope` is shared infrastructure.** Changes to the envelope affect all events and require platform team approval.
4. **Consumers must not dictate producer schema.** If a consumer needs additional data, the consumer team requests it from the producer team — the producer team decides how to expose it.

### Change workflow

```
1. Developer opens PR with .proto changes
2. `make proto-gen` runs (CI verifies generated code matches)
3. (Recommended) `buf breaking` runs in CI to detect incompatible changes
4. Producer team lead reviews
5. Consumer team lead reviews (required for shared schemas)
6. Merge → deploy consumer first → deploy producer
```

## CI Recommendations

To enforce schema compatibility automatically, we recommend adding **buf** to CI:

```yaml
# .github/workflows/proto-check.yml (example)
- name: Check Protobuf breaking changes
  uses: bufbuild/buf-action@v1
  with:
    input: pkg/infra/events/schemas
    against: 'https://github.com/leet-gaming/match-making-api.git#branch=develop,subdir=pkg/infra/events/schemas'
```

This catches:
- Removed fields
- Changed field numbers
- Changed field types
- Renamed messages

Without needing a Schema Registry.

## Compilation

```bash
make proto-gen
```

Requires `protoc` and `protoc-gen-go` installed. The `install-tools` target installs the Go plugin.

## References

- [CloudEvents 1.0 Specification](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md)
- [CloudEvents JSON Format](https://github.com/cloudevents/spec/blob/main/cloudevents/formats/json-format.md)
- [ADR-001: Schema Registry Decision](ADR-001-schema-registry-decision.md)
- [Protobuf Language Guide — Updating a Message Type](https://protobuf.dev/programming-guides/proto3/#updating)
- [buf — Protobuf breaking change detection](https://buf.build/docs/breaking/overview)
- Epic §9 — Event Payloads with Resource Ownership
- Epic §10 — Event Schemas, Schema Evolution
- Issue #16 — Event versioning
- Issue #34 — Distributed tracing (correlation_id)
