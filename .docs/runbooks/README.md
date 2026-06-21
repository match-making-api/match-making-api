# Matchmaking Operations Runbooks (2505-004)

On-call runbooks for matchmaking Kafka and queue operations.

| Runbook | Alert / trigger |
|---------|-----------------|
| [Consumer lag spike](./consumer-lag-spike.md) | `MatchmakingConsumerLagHigh`, lag > 1000 |
| [DLQ filling up](./dlq-filling-up.md) | Messages in `matchmaking.dlq`, DLQ depth > 0 |
| [Queue depth threshold](./queue-depth-threshold.md) | Pool size / Redis queue depth above SLO |

## Ownership

| Area | Team | Channel |
|------|------|---------|
| Kafka consumers | Matchmaking platform | `#matchmaking-ops` |
| replay-api producers | Replay API | `#replay-api` |
| Dashboards / alerts | SRE + Matchmaking | Grafana folder `Matchmaking Kafka` |

## Related docs

- Dashboard: `.docs/MONITORING_DASHBOARD.md` (2505-001)
- DLQ: `.docs/DLQ_ERROR_HANDLING.md` (2505-003)
- Tracing: `.docs/DISTRIBUTED_TRACING.md` (2505-002)
