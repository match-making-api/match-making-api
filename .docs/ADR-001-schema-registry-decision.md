# ADR-001: Schema Registry — Decision Record

| Field       | Value |
|-------------|-------|
| **Status**  | Accepted |
| **Date**    | 2026-02-17 |
| **Deciders** | Match-Making-API team |
| **Related** | Epic §10 Event Schemas, Issue #16 (event versioning), Issue #19 (this story) |

## Context

The matchmaking platform exchanges events between **replay-api** (producer) and **match-making-api** (consumer) over Kafka. Events use **Protobuf** schemas serialised with `protojson` (JSON wire format). The epic references "event versioning" and "schema evolution" and mentions the possibility of a **Schema Registry** (e.g. Confluent Schema Registry, Karapace).

We need to decide whether to deploy a Schema Registry or rely on repository-based schemas with in-envelope versioning.

## Decision

**We will NOT adopt a Schema Registry.** Instead, we use **repo-based Protobuf schemas** with a `dataschema_version` field in the event envelope for versioning.

## Rationale

### Why repo-based schemas are sufficient

1. **Protobuf has built-in evolution guarantees.** Field numbers are immutable, new fields are additive, and `reserved` prevents accidental reuse of removed fields. These rules are enforced at compile time by `protoc`.

2. **`dataschema_version` in the envelope.** Every `MatchmakingEvent` carries an `EventEnvelope` with a `dataschema_version` (int32) that consumers inspect to decide how to deserialise the payload. This gives us explicit version routing without an external service.

3. **`protojson` wire format.** We serialise with `protojson`, producing human-readable JSON. A Schema Registry typically adds value for binary formats (Avro, binary Protobuf) where the schema is needed to deserialise — with JSON, the schema is already implicit in the field names.

4. **Operational simplicity.** A Schema Registry is another service to deploy, monitor, back up, and secure. At our current scale (2 services, <10 event types), the operational cost exceeds the benefit.

5. **Code review as compatibility gate.** Proto file changes go through pull request review. CI can run `buf breaking` or `protoc` diffing to catch incompatible changes automatically — this provides the same safety net a registry would.

### When to reconsider

Revisit this decision if any of the following become true:

- **>5 independent services** produce or consume matchmaking events (schema discovery becomes hard)
- **Binary Protobuf** wire format is adopted (consumers need the schema to deserialise)
- **Avro** schemas are introduced (Avro requires a schema at read time)
- **External partners** consume events and need a self-service schema catalog
- **Schema evolution violations** slip through code review repeatedly

## Consequences

### Positive

- No additional infrastructure to deploy or maintain
- Schema changes are versioned alongside application code (single source of truth)
- `protoc` + CI catch breaking changes at build time
- Lower cognitive overhead for developers (no registry API, no subject naming strategy)

### Negative

- Schema discovery relies on reading the repo — no self-service catalog UI
- No automatic compatibility enforcement at produce time (must rely on CI + code review)
- If we move to binary Protobuf later, consumers need access to the `.proto` files to deserialise

### Neutral

- The `dataschema_version` field is already implemented and in use
- `EVENT_SCHEMAS.md` serves as the living schema catalog
- Topic → schema mapping is documented and maintained manually

## Alternatives Considered

### Confluent Schema Registry

- **Pros**: Industry standard, automatic compatibility checks, schema catalog UI, supports Avro/Protobuf/JSON Schema
- **Cons**: Java-based, requires Kafka for storage (or CP-compatible backend), operational overhead, licensing concerns (Confluent Community License)
- **Verdict**: Overkill for our scale; revisit if service count grows significantly

### Karapace (open-source, Aiven)

- **Pros**: Confluent-compatible API, Python-based (lighter), truly open-source (Apache 2.0)
- **Cons**: Still an additional service to deploy/monitor; less mature than Confluent
- **Verdict**: Better license story, but same operational overhead argument applies

### buf.build (schema linting + breaking change detection in CI)

- **Pros**: No runtime service, runs in CI, excellent Protobuf support, free tier
- **Cons**: Only checks at CI time, no runtime schema negotiation
- **Verdict**: Complementary to our current approach — **recommended as a future CI enhancement** without needing a full registry

## Related Documents

- [EVENT_SCHEMAS.md](.docs/EVENT_SCHEMAS.md) — schema catalog, topic mapping, evolution rules
- [KAFKA_SECURITY.md](.docs/KAFKA_SECURITY.md) — Kafka security configuration
- `pkg/infra/events/schemas/matchmaking_events.proto` — canonical Protobuf definitions
