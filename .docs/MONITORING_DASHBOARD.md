# Event Monitoring Dashboard (2505-001)

Operational guide for matchmaking Kafka metrics, dashboards, and alerts.

## Access

| Component | Default | Env vars |
|-----------|---------|----------|
| App metrics | `http://<pod>:9090/metrics` | `METRICS_ENABLED` (default on), `METRICS_PORT` (default `9090`) |
| Kafka Exporter | `http://kafka-exporter-matchmaking:9308/metrics` | Deploy: `deploy/monitoring/kafka-exporter.yaml` |
| Grafana dashboard | Import `deploy/monitoring/grafana-dashboard-matchmaking.json` | Datasource: Prometheus |

## Key metrics

| Metric | Meaning |
|--------|---------|
| `kafka_consumergroup_lag` | Consumer lag per group/topic (from Kafka Exporter) |
| `matchmaking_kafka_messages_consumed_total` | Messages consumed by topic and group |
| `matchmaking_kafka_messages_produced_total` | Messages produced by topic |
| `matchmaking_kafka_consumer_errors_total` | Consumer fetch/process failures |
| `matchmaking_kafka_producer_errors_total` | Producer publish failures |
| `matchmaking_kafka_message_processing_seconds` | Handler latency histogram |

## Alerts

Rules live in `deploy/monitoring/prometheus-alerts.yaml`.

| Alert | Default threshold | Tuning |
|-------|-------------------|--------|
| `MatchmakingConsumerLagHigh` | lag > 1000 for 5m | Edit `expr` or use Alertmanager inhibition |
| `MatchmakingConsumerGroupDown` | 0 members for 2m | Verify group name regex |
| `MatchmakingProducerErrorRateHigh` | > 0.1/s for 5m | Adjust rate window |
| `MatchmakingConsumerErrorRateHigh` | > 0.1/s for 5m | Adjust rate window |

Route alerts via Alertmanager to Slack/PagerDuty (team channel: `#matchmaking-ops`).

## Dashboard panels

The Grafana JSON includes:

1. Consumer lag by group (matchmaking topics)
2. Produce/consume throughput (rates from app + exporter)
3. Error rates (producer/consumer)
4. Processing latency p50/p95/p99

## Runbooks

See `.docs/runbooks/` for response steps when alerts fire.
