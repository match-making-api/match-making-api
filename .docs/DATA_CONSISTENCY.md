# Data Consistency Monitoring (2509-002)

During Strimzi Migration Phase 2 dual-write, this service periodically compares **legacy** matchmaking sessions (replay-api Mongo `matchmaking_sessions`) with the **new** path active queue (`ActiveQueueStore` / Redis).

## What is compared

| Entity | Legacy source | New source | Drift kinds |
|--------|---------------|------------|-------------|
| Queue membership | sessions with status `queued` / `searching` | `ActiveQueueStore` | `legacy_only_queue`, `new_only_queue` |
| Match state | sessions with status `matched` | still present in `ActiveQueueStore` | `matched_still_queued` |

## Worker

```text
cmd/workers/consistency-check
```

Runs `ConsistencyTicker` on a fixed interval (default 60s). Metrics are exposed on the process metrics port (`METRICS_PORT`, default 9090).

## Feature flags / env

| Env | Default | Effect |
|-----|---------|--------|
| `MATCHMAKING_CONSISTENCY_CHECK_ENABLED` | `true` | When `false`, ticker no-ops. |
| `MATCHMAKING_CONSISTENCY_CHECK_INTERVAL` | `60s` | Check period (Go duration). |
| `MATCHMAKING_CONSISTENCY_DRIFT_ALERT_THRESHOLD` | `1` | Drift count that escalates alert log level. |
| `MATCHMAKING_LEGACY_SESSIONS_COLLECTION` | `matchmaking_sessions` | Legacy Mongo collection. |
| `MATCHMAKING_LEGACY_SESSIONS_DB` | (empty) | Optional DB override; else `MONGO_DB_NAME`. |

## Metrics

| Metric | Meaning |
|--------|---------|
| `matchmaking_consistency_checks_total{outcome}` | Check runs (`ok`, `drift`, `error`) |
| `matchmaking_consistency_drift_total{kind}` | Cumulative drift findings |
| `matchmaking_consistency_drift_current{kind}` | Latest drift gauge |
| `matchmaking_consistency_last_check_timestamp_seconds` | Last check time |
| `matchmaking_consistency_check_duration_seconds` | Check latency |

## Alerts and runbook

Prometheus rules: `deploy/monitoring/prometheus-alerts.yaml` (`MatchmakingConsistencyDrift`, `MatchmakingConsistencyCheckFailed`, `MatchmakingConsistencyCheckStale`).

Response steps: `.docs/runbooks/consistency-alert.md`.

## Related

- Dual-write: `.docs/DUAL_WRITE.md` (2509-001)
- Rollback: `.docs/runbooks/dual-write-rollback.md`
- Dashboard: `.docs/MONITORING_DASHBOARD.md`
