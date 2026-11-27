# Runbook: Legacy decommission rollback (2509-004)

**Trigger:** Critical regression after Phase 4 new-path-only cutover  
**Severity:** Sev-1 / Sev-2 depending on match-create failure rate

## Goal

Temporarily restore dual-write / legacy-capable operation (or hard-stop the new path) until the regression is fixed, then return to Phase-4 defaults.

## Immediate steps

1. Confirm blast radius: Kafka lag, MatchCreated produce errors, queue join success rate, DLQ depth.
2. On **replay-api** (producer enforcement):
   - Set `KAFKA_MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=0` for full legacy cohort, **or**
   - Set `KAFKA_MATCHMAKING_PRODUCER_ENABLED=false` for hard stop of Kafka produce.
3. On **match-making-api** (mirror + consumer):
   - Mirror `MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=0`.
   - Optionally set `MATCHMAKING_DUAL_WRITE_ENABLED=false` so handlers no-op if residual Kafka traffic arrives.
   - Optionally set `MATCHMAKING_DUAL_WRITE_FAILURE_POLICY=legacy_only_and_alert` if soft produce failures are preferred during triage.
4. Redeploy / reload both services; confirm gauges and startup logs.
5. If product requires legacy session restore: follow replay-api backup restore SOP (out of this repo).

## Optional consistency window

Set `MATCHMAKING_CONSISTENCY_CHECK_ENABLED=true` and run `cmd/workers/consistency-check` while dual paths coexist again. Follow [consistency-alert.md](./consistency-alert.md).

## Return to Phase 4

1. Fix root cause on the new path.
2. Restore defaults: traffic `100`, dual-write/new-path `enabled=true`, failure policy `fail_request`, consistency `enabled=false`.
3. Validate join → Kafka → pool → MatchCreated for a canary cohort before full traffic.

## Related

- Decommission doc: `.docs/LEGACY_DECOMMISSION.md`
- Dual-write rollback: [dual-write-rollback.md](./dual-write-rollback.md)
- Traffic rollout: [traffic-migration-rollout.md](./traffic-migration-rollout.md)
