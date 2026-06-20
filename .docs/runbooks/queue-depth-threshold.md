# Runbook: Queue depth / pool size above threshold

**Trigger:** Active queue depth or pool size exceeds configured threshold (e.g. Redis `ActiveQueueStore`, matchmaking pool metrics)  
**Severity:** Warning  
**Owner:** Matchmaking platform (`#matchmaking-ops`)

## Symptoms

- Players report long queue times / no matches forming.
- Elevated `total_in_queue` in `QUEUE_STATUS_UPDATED` WebSocket payloads.
- Redis keys for active queues growing without draining.
- Match creation rate drops while join rate stays high.

## Checks

1. **Dashboard** — throughput panels (2505-001): produce/consume on `matchmaking.commands` and `matchmaking.matches.created`.
2. **Consumers** — `matchmaking-commands` and match-creation worker running.
3. **Matching logic** — insufficient players per region/game mode; MMR spread too narrow.
4. **replay-api** — join API errors preventing leaves (stale pool entries).
5. **Tracing** (2505-002) — slow path on `PlayerQueued` handler.

## Mitigation

| Scenario | Action |
|----------|--------|
| Consumer down | Restart/scale `cmd/consumers/matchmaking-commands` |
| Redis full/slow | Check Dragonfly/Redis memory; evict stale queues per runbook |
| Config too strict | Review game mode min players / MMR band (product approval) |
| Surge traffic | Scale consumers; communicate ETA to support |

## Escalation

| Condition | Action |
|-----------|--------|
| Depth rising > 30m with healthy consumers | Engage product for match rules review |
| Regional outage | Failover or disable queue for affected region (feature flag) |

## Links

- Queue flows: `.docs/MATCH_CREATION_FLOW.md`
- Monitoring: `.docs/MONITORING_DASHBOARD.md`
- Lag runbook: [consumer-lag-spike.md](./consumer-lag-spike.md)
