# Legacy Decommission & Cleanup (2509-004)

Strimzi Migration **Phase 4**: traffic is fully on the Kafka + match-making-api path. Dual-write is retired as the operating mode; legacy routing and consistency checks are off by default.

## Steady-state (new path only)

| Concern | Phase 4 default | Notes |
|---------|-----------------|-------|
| New-path handlers | `MATCHMAKING_DUAL_WRITE_ENABLED=true` | `false` = emergency kill switch only |
| MatchCreated failure | `MATCHMAKING_DUAL_WRITE_FAILURE_POLICY=fail_request` | Soft `legacy_only_and_alert` retired from defaults |
| Traffic percent | `MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=100` | Mirror of producer; legacy cohort retired |
| Consistency worker | `MATCHMAKING_CONSISTENCY_CHECK_ENABLED=false` | Legacy session comparison not needed in steady state |

Producer (replay-api) must keep `KAFKA_MATCHMAKING_PRODUCER_ENABLED=true` and traffic percent `100` so all joins/leaves reach Kafka.

## Cleanup checklist

1. Confirm 2509-003 at 100% new path with stakeholder sign-off.
2. Deploy match-making-api with Phase-4 defaults (this story).
3. Align replay-api producer flags to 100% new path / producer enabled.
4. Stop scheduling the consistency-check worker in steady state (or leave it with enabled=false).
5. Archive or freeze dual-write / traffic-migration runbooks as rollback references.
6. Do **not** invent or delete Kafka topic names in this story; topic teardown is a separate infra change if ever required.
7. Do **not** drop Mongo legacy session collections from this service; archival is ops-owned if product requires it.
8. Confirm wallet untouched: this service only publishes chargeable events elsewhere; no balance deduction.

## Event-flow optimizations (documented)

| Before (dual-write / gradual) | After (Phase 4) | Benefit |
|-------------------------------|-----------------|---------|
| Handlers could no-op when dual-write disabled | New path always on in steady state | No silent skip of queue traffic |
| `legacy_only_and_alert` soft-failed MatchCreated | Default `fail_request` | Faster fail / retry; no unmatched soft pairs |
| Consistency ticker compared legacy Mongo vs Redis | Disabled by default | Less Mongo read load and alert noise |
| Traffic percent < 100 routed legacy cohort | Locked at 100% | Single path; simpler ops and fewer A/B branches |

## Rollback plan

See [runbooks/legacy-decommission-rollback.md](./runbooks/legacy-decommission-rollback.md).

Summary:

1. Lower producer + mirror traffic percent toward `0` (or set producer enabled false).
2. Optionally set `MATCHMAKING_DUAL_WRITE_ENABLED=false` to stop new-path handlers.
3. Optionally re-enable consistency checks during the rollback window.
4. Restore legacy session handling on replay-api from backup / prior deployment if product still supports it.

## Related

- Dual-write history: `.docs/DUAL_WRITE.md` (2509-001)
- Consistency: `.docs/DATA_CONSISTENCY.md` (2509-002)
- Traffic: `.docs/TRAFFIC_MIGRATION.md` (2509-003)
- Epic: Strimzi Migration Phase 4
