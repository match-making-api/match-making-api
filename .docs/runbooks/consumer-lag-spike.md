# Runbook: Kafka consumer lag spike (matchmaking topics)

**Alert:** `MatchmakingConsumerLagHigh`  
**Severity:** Critical  
**Owner:** Matchmaking platform (`#matchmaking-ops`)

## Detection

1. Grafana dashboard **Matchmaking Kafka** → panel *Consumer Lag*.
2. Prometheus: `kafka_consumergroup_lag{consumergroup=~"matchmaking-.*"} > 1000` for 5m.
3. Kafka Exporter metrics on `:9308/metrics`.

## Common causes

| Cause | Signal |
|-------|--------|
| Consumer pod crash / scale to zero | `kafka_consumergroup_members == 0` |
| Slow handler (DB, external API) | High `matchmaking_kafka_message_processing_seconds` p95 |
| Traffic spike from replay-api | High produce rate on `matchmaking.commands` |
| Broker or network issue | Fetch errors in consumer logs |

## Mitigation steps

1. **Confirm scope** — note `consumergroup`, `topic`, lag value.
2. **Check consumer health** — `kubectl get pods -l app=<consumer>`; restart if CrashLoopBackOff.
3. **Scale consumers** — increase replicas for the affected `cmd/consumers/*` deployment (same `group.id`).
4. **Check downstream** — MongoDB/Redis latency; fix infra if degraded.
5. **Temporary threshold** — if expected burst (tournament), tune alert in `deploy/monitoring/prometheus-alerts.yaml` (requires change review).

## Escalation

| Condition | Action |
|-----------|--------|
| Lag not decreasing in 15m after scale-up | Page matchmaking on-call |
| Multiple groups lagging | Escalate to Kafka/SRE |
| Suspected bad deploy | Roll back affected consumer deployment |

## Links

- Dashboard: `.docs/MONITORING_DASHBOARD.md`
- Deploy: `deploy/monitoring/grafana-dashboard-matchmaking.json`
