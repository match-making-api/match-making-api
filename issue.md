## Context

Event **schema versioning** (#16) can be backed by a **Schema Registry** (e.g. Confluent, Karapace) for Avro/Protobuf. The epic references **event versioning** and **schema evolution**; a registry centralises schemas and enables compatibility checks. This story implements **Schema Registry** setup and **integration** with matchmaking producers/consumers **if** the team adopts it.

## Description

**Set up Schema Registry** for matchmaking events (if chosen): deploy registry (e.g. alongside Strimzi), **register** schemas for PlayerQueued, MatchCreated, and other matchmaking events (#16), and **configure** producers/consumers to use it. Document **topic–schema** mapping, **compatibility** rules (e.g. backward), and **evolution** process. If **not** using a registry, document the decision and **alternative** (e.g. schemas in repo, version in envelope).

## Background

- Epic §10 Event Schemas, § Event Schema Evolution Strategy; #16 **event versioning**.
- Schema Registry: store Avro/Protobuf schemas; producers/consumers fetch by topic or ID; compatibility checks on register.
- **Optional**: some teams use repo-based schemas only; this story is **conditional** on adopting a registry.

## Scope

**In scope:** Schema Registry deployment (if used); registration of matchmaking schemas; producer/consumer config; **topic–schema** mapping and **compatibility** rules; **documentation** (or “not used” decision).

**Out of scope:** Defining new event schemas (#16); Kafka cluster (#15).

## Acceptance Criteria

- [ ] **Decision** documented: Schema Registry **in use** or **not**; if not, **alternative** (e.g. repo + version field) documented.
- [ ] If **in use**: Registry **deployed**; matchmaking schemas **registered**; producers/consumers **configured**; **topic–schema** mapping and **compatibility** rules documented.
- [ ] **Evolution** process (add field, break change) documented; **ownership** for schema changes clear.

## Dependencies

- #15 (Kafka), #16 (schemas); choice of registry (Confluent, Karapace, etc.) or “no registry.”

## References

- Epic §10 Event Schemas, Schema Evolution; #16.
