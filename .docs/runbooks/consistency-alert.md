# Runbook: Consistency drift / check failure (2509-002)

**Trigger:** `MatchmakingConsistencyDrift`, `MatchmakingConsistencyCheckFailed`, or `MatchmakingConsistencyCheckStale`  
**Severity:** Warning (drift) / Critical (check failed or stale)  
**Owner:** Matchmaking platform (`#matchmaking-ops`)

## Symptoms

- Logs with `alert=MatchmakingConsistencyDrift` or `MatchmakingConsistencyCheckFailed`.
- Grafana panels: non-zero `matchmaking_consistency_drift_current`.
- `matchmaking_consistency_last_check_timestamp_seconds` not updating.

## Investigate

1. **Dashboard** — consistency panels in Matchmaking Grafana; note drift `kind` (`legacy_only_queue`, `new_only_queue`, `matched_still_queued`).
2. **Worker** — confirm `cmd/workers/consistency-check` is running; check pod restarts.
3. **Dual-write flags** — `MATCHMAKING_DUAL_WRITE_ENABLED` on match-making-api; `KAFKA_MATCHMAKING_PRODUCER_ENABLED` on replay-api.
4. **Kafka lag** — `matchmaking.commands` / `matchmaking.matches.created` consumer lag (2505-001).
5. **Sample player IDs** — from alert log samples; compare Mongo `matchmaking_sessions` vs Redis active queue entry.

## Drift meaning

| Kind | Meaning | Typical cause |
|------|---------|---------------|
| `legacy_only_queue` | Legacy session queued; not in new active queue | Kafka produce skipped/failed; dual-write disabled; consumer lag |
| `new_only_queue` | In new active queue; no legacy queued/searching session | Stale Redis entry; leave not produced; session expired/cancelled in legacy |
| `matched_still_queued` | Legacy matched; still in new active queue | MatchCreated consumed by legacy but pool leave missed |

## Repair

| Scenario | Action |
|----------|--------|
| Transient lag | Wait for consumers to catch up; re-check gauges |
| Producer off / dual-write off | Re-enable flags if unintended; otherwise expected during rollback |
| Persistent `legacy_only_queue` | Verify replay-api produce path; replay failed commands carefully if DLQ has them |
| Persistent `new_only_queue` / `matched_still_queued` | Remove stale ActiveQueue entries only after confirming legacy truth; do not invent balance/wallet changes |
| Check cannot read Mongo/Redis | Fix connectivity; do not raise traffic (2509-003) until checks are green |

## Escalate / rollback

1. If drift grows for >15m under dual-write: engage `#matchmaking-ops` + `#replay-api`.
2. If impact on player joins/matches: follow [dual-write-rollback.md](./dual-write-rollback.md) (disable new path).
3. Do **not** start traffic migration (2509-003) while consistency alerts are firing.

## Links

- Consistency doc: `.docs/DATA_CONSISTENCY.md`
- Dual-write: `.docs/DUAL_WRITE.md`
- Monitoring: `.docs/MONITORING_DASHBOARD.md`
