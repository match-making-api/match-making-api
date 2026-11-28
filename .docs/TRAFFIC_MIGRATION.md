# Traffic Migration — Gradual Rollout / A/B (2509-003)

Strimzi Migration **Phase 3**: route a configurable **percentage** of matchmaking traffic to the **new** Kafka + match-making-api path vs **legacy** (replay-api sessions only), support **A/B** assignment, and **roll back** by setting percent to `0`.

> **Phase 4 (2509-004):** Steady state is **100% new path**. Percent knobs remain for emergency rollback only. See `.docs/LEGACY_DECOMMISSION.md`.

## Routing ownership

| Layer | Role |
|-------|------|
| **replay-api producer** | Enforces split: only the new-path cohort publishes `PlayerQueued` (and related) to Kafka. |
| **match-making-api** | Mirrors config for ops visibility, exposes metrics, documents runbook; processes whatever arrives on the new path. |

Topic names are unchanged (`matchmaking.commands`, `matchmaking.matches.created`, etc.). This story does not invent topics.

## Feature flags / env

### match-making-api (mirror + metrics)

| Env | Default | Effect |
|-----|---------|--------|
| `MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT` | `100` | Expected rollout percent (0 = rollback, 100 = full). Synced to gauge. |
| `MATCHMAKING_TRAFFIC_STRATEGY` | `percent` | `percent` = sticky hash by player id; `region` = region allowlist. |
| `MATCHMAKING_TRAFFIC_REGION_ALLOWLIST` | (empty) | Comma-separated regions when strategy is `region`. |

### replay-api (enforcement)

| Env | Default | Effect |
|-----|---------|--------|
| `KAFKA_MATCHMAKING_PRODUCER_ENABLED` | `true` | Master switch; `false` disables all Kafka produce (legacy only). |
| `KAFKA_MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT` | `100` | 0–100 sticky cohort size for new-path produce. |
| `KAFKA_MATCHMAKING_TRAFFIC_STRATEGY` | `percent` | `percent` \| `region`. |
| `KAFKA_MATCHMAKING_TRAFFIC_REGION_ALLOWLIST` | (empty) | Regions that receive new-path produce when strategy is `region`. |

Algorithm (both sides): `fnv32a(player_id) % 100 < percent` for sticky `percent` strategy.

## Gradual rollout plan

1. Preconditions: dual-write healthy (2509-001), consistency checks green (2509-002).
2. Set percent to **10** on replay-api (+ mirror on match-making-api). Redeploy / reload.
3. Validate join → Kafka → pool → match for the new-path cohort; watch error and drift metrics.
4. Increase **25 → 50 → 75 → 100** after each validation window.
5. Success criteria: produce/consume rates scale with percent; no sustained consistency drift; match quality (errors, lag, drift) acceptable vs legacy cohort.

## Rollback

Set `KAFKA_MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=0` (and mirror `MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT=0`). Optionally set `KAFKA_MATCHMAKING_PRODUCER_ENABLED=false` / `MATCHMAKING_DUAL_WRITE_ENABLED=false` for hard new-path off. See [runbooks/traffic-migration-rollout.md](./runbooks/traffic-migration-rollout.md).

## Metrics (quality comparison)

| Metric | Meaning |
|--------|---------|
| `matchmaking_traffic_new_path_percent` | Configured rollout percent |
| `matchmaking_traffic_route_decisions_total{path}` | Decisions `new` vs `legacy` (when router is invoked) |
| `matchmaking_kafka_messages_produced_total` / `_consumed_total` | New-path volume |
| `matchmaking_consistency_drift_current` | Drift while dual-write + partial traffic (2509-002) |

Compare new-path Kafka error/lag rates against legacy session success while percent < 100.

## Related

- Dual-write: `.docs/DUAL_WRITE.md` (2509-001)
- Consistency: `.docs/DATA_CONSISTENCY.md` (2509-002)
- Dashboard: `.docs/MONITORING_DASHBOARD.md`
- Router: `pkg/infra/traffic`
