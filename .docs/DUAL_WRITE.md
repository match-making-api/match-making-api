# Dual-Write for Matchmaking Operations (2509-001)

Strimzi Migration **Phase 2**: the same logical operation updates the **legacy** path and the **Kafka + match-making-api** path so traffic can cut over and roll back without data loss.

## Operations

| Operation | Legacy path | New path | Dual-write owner |
|-----------|-------------|----------|------------------|
| Queue join | replay-api matchmaking session (Mongo) | `PlayerQueued` → `matchmaking.commands` → pool | replay-api producer + this consumer |
| Queue leave | replay-api session cancel + legacy queue event | `PlayerLeftQueue` → `matchmaking.commands` → pool remove | replay-api producer + this consumer |
| Match create | pair/pool state in match-making-api; replay-api session update via event | `MatchCreated` on `matchmaking.matches.created` | this service (state + produce) |

Producer enable flag (replay-api): `KAFKA_MATCHMAKING_PRODUCER_ENABLED` (default `true`).

## Feature flags (match-making-api)

| Env | Default | Effect |
|-----|---------|--------|
| `MATCHMAKING_DUAL_WRITE_ENABLED` | `true` | When `false`, `PlayerQueued` / `PlayerLeftQueue` handlers no-op (new path off — rollback). |
| `MATCHMAKING_DUAL_WRITE_FAILURE_POLICY` | `legacy_only_and_alert` | `legacy_only_and_alert`: keep local pair, log/alert if `MatchCreated` produce fails. `fail_request`: return error and fail the consumer handler. |

## Consistency model

- **Join:** legacy session commit is authoritative for the player-facing API; Kafka produce failure on replay-api is currently non-fatal (warn) so legacy-only may occur until produce succeeds or is reconciled.
- **Leave:** same pattern once `PlayerLeftQueue` is produced (see cross-repo RA-2502-003 if producer still open).
- **Match create:** pair is persisted first; `MatchCreated` is the new-path signal to replay-api. With `legacy_only_and_alert`, a produce failure leaves an unmatched pair that may need reconciliation (2509-002).
- **Ordering:** commands share `matchmaking.commands`; consumers process in offset order per partition.

Monitoring of drift between legacy sessions and pool state is **out of scope** here (2509-002).

## Rollback

See [runbooks/dual-write-rollback.md](./runbooks/dual-write-rollback.md).

## Related

- Epic: Strimzi Migration Phase 2
- Consumer: `cmd/consumers/matchmaking-commands`
- Config: `pkg/infra/config/dual_write.go`
