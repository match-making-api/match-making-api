# Event Schemas - Matchmaking

Definition of canonical schemas for matchmaking events exchanged via Kafka between **replay-api** and **match-making-api**. The schemas ensure agreement on payload structure, field types, and resource ownership metadata.

## Schema Location

- **Proto files**: `pkg/infra/events/schemas/matchmaking_events.proto`
- **Generated Go code**: `pkg/infra/events/schemas/matchmaking_events.pb.go`
- **Documentation**: `.docs/EVENT_SCHEMAS.md` (this file)

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

### MatchCreated (match-making-api → replay-api)

Emitted when a match is created.

**Payload:**
- `match_id`, `lobby_id`, `tenant_id`, `client_id` (required)
- `players[]`: `player_id`, `party_id`, `resource_permissions`
- `game_server`: `server_id`, `region`, `resource_owner_id`

### MatchCompleted (optional)

Placeholder for producer use (2501-003). Payload includes `match_id`, `player_ids`, `winner_team_id`, `is_draw`, etc.

### RatingsUpdated (optional)

Placeholder for the epic. Payload includes `deltas[]` of MMR per player.

## Versioning Strategy

### `dataschema_version` field in the envelope

- The `dataschema_version` (int32) field in the envelope indicates the payload schema version.
- Initial version: `1`.
- Consumers should support multiple versions when possible and document incompatibilities.

### Protobuf evolution rules

1. **Do not remove fields**: use `reserved` if you need to deprecate a field.
2. **New fields**: always add with a new field number; optional fields or fields with default values are safe.
3. **New payload types**: add to the `oneof` in `MatchmakingEvent` with a new number.
4. **Renaming**: Protobuf uses field numbers; renaming the field name is safe (JSON maintains compatibility via `json_name`).

### Where schemas are stored

- **Repository**: schemas live in `pkg/infra/events/schemas/` in the repo.
- **Schema Registry** (if adopted): document topic → schema mapping and the process for registering new versions in Confluence or the team README.

## Topic → Schema Mapping

| Topic | Events | Schema |
|-------|--------|--------|
| `matchmaking.commands` | PlayerQueued | `PlayerQueuedPayload` |
| `matchmaking.matches.created` | MatchCreated | `MatchCreatedPayload` |
| `matchmaking.matches` | MatchCompleted | `MatchCompletedPayload` |
| (TBD) | RatingsUpdated | `RatingsUpdatedPayload` |

## How to add new events

1. Add the payload message in `matchmaking_events.proto`.
2. Add the payload to the `oneof` in `MatchmakingEvent`.
3. Run `make proto-gen` to regenerate the Go code.
4. Update this document with the new event and topic mapping.
5. If using Schema Registry, register the new schema version.

## How to evolve existing events

1. Add new fields with unused field numbers.
2. For optional fields, use `optional` in proto3.
3. Increment `dataschema_version` in the envelope when making incompatible changes (e.g. field removal).
4. Document supported versions and the deprecation plan.
5. Run `make proto-gen` and update consumers/producers.

## Compilation

```bash
make proto-gen
```

Requires `protoc` and `protoc-gen-go` installed. The `install-tools` target installs the Go plugin.

## References

- [CloudEvents 1.0 Specification](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md)
- [CloudEvents JSON Format](https://github.com/cloudevents/spec/blob/main/cloudevents/formats/json-format.md)
- Epic §9 — Event Payloads with Resource Ownership
- Epic §10 — Event Schemas, Schema Evolution
- Issue #34 — Distributed tracing (correlation_id)
