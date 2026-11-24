# Runbook: Traffic migration rollout / rollback (2509-003)

**Trigger:** Planned Phase 3 gradual cutover, A/B quality review, or emergency rollback to legacy  
**Severity:** High when raising percent under active drift/errors  
**Owner:** Matchmaking platform (`#matchmaking-ops`) + Replay API (`#replay-api`)

## Goal

Route a configurable **percentage** of matchmaking traffic to the **new** Kafka / match-making-api path, validate quality, increase gradually (10% → 100%), or **rollback** to 0% (legacy).

## Preconditions

1. Dual-write enabled and stable (2509-001).
2. Consistency checks green — no firing drift alerts (2509-002).
3. Dashboards for produce/consume, lag, errors available (2505-001).

## Gradual increase

1. Set on **replay-api**: `KAFKA_MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=10` (keep `KAFKA_MATCHMAKING_PRODUCER_ENABLED=true`).
2. Mirror on **match-making-api**: `MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=10` and restart consumers so the gauge updates.
3. Optional A/B by region: `KAFKA_MATCHMAKING_TRAFFIC_STRATEGY=region` and `KAFKA_MATCHMAKING_TRAFFIC_REGION_ALLOWLIST=us-east` (mirror `MATCHMAKING_TRAFFIC_*`).
4. Validation window (suggested ≥ 30m):
   - New-path joins produce `PlayerQueued` and land in pool.
   - `matchmaking_traffic_new_path_percent` == configured value.
   - Consumer lag and error rates within SLO.
   - Consistency drift not sustained above alert threshold.
5. Raise percent: 25 → 50 → 75 → 100, repeating step 4 each time.

## Immediate rollback (0% / legacy)

1. Set `KAFKA_MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=0` on replay-api and redeploy / reload.
2. Mirror `MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=0` on match-making-api.
3. Confirm Kafka produce rate for matchmaking commands drops toward residual dual-write only / zero new cohort.
4. If hard stop needed: follow [dual-write-rollback.md](./dual-write-rollback.md) (`PRODUCER_ENABLED=false`, `DUAL_WRITE_ENABLED=false`).

## Checks

1. Gauge `matchmaking_traffic_new_path_percent`.
2. Counter `matchmaking_traffic_route_decisions_total{path="new|legacy"}` (when router is exercised).
3. Produce/consume rates scale with percent.
4. Consistency worker (2509-002) remains healthy before next increase.

## Links

- Traffic doc: `.docs/TRAFFIC_MIGRATION.md`
- Dual-write rollback: [dual-write-rollback.md](./dual-write-rollback.md)
- Consistency: [consistency-alert.md](./consistency-alert.md)
- Monitoring: `.docs/MONITORING_DASHBOARD.md`
