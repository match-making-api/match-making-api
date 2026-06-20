# Runbook: DLQ filling up

**Alert:** DLQ depth > 0 or ingest rate on `matchmaking.dlq`  
**Severity:** Warning → Critical if sustained  
**Owner:** Matchmaking platform (`#matchmaking-ops`)

## Detection

1. Kafka Exporter lag/offset on topic `matchmaking.dlq`.
2. Consumer logs: `Message sent to DLQ after retries` (2505-003).
3. Grafana error panels + manual `kafka-console-consumer` on DLQ topic.

## Inspect

```bash
# Sample DLQ messages (adjust brokers)
kafka-console-consumer --bootstrap-server $BROKERS \
  --topic matchmaking.dlq --from-beginning --max-messages 5
```

Parse JSON fields: `original_topic`, `error`, `retry_count`, `payload`.

## Root cause analysis

| Error pattern | Likely cause |
|---------------|--------------|
| Unmarshal / schema | Producer schema mismatch; check proto version |
| Resource ownership | Invalid tenant/client on command |
| DB timeout | MongoDB capacity or index issue |
| Transient external | Retry may succeed |

## Retry (manual)

**Only after fixing root cause or confirming transient failure.**

```bash
cd match-making-api
go run ./cmd/tools/dlq-retry -brokers=$BROKERS -limit=10 -dry-run=true
# then
go run ./cmd/tools/dlq-retry -brokers=$BROKERS -limit=10 -dry-run=false
```

See `.docs/DLQ_ERROR_HANDLING.md` for idempotency requirements.

## Alert tuning

- Warning: first message in DLQ in 1h.
- Critical: DLQ rate > N/min for 10m.

## Escalation

| Condition | Action |
|-----------|--------|
| Poison message repeating | Disable auto-retry; fix handler; skip offset with approval |
| DLQ > 1000 messages | Page on-call; pause non-critical consumers if needed |

## Links

- DLQ format & policy: `.docs/DLQ_ERROR_HANDLING.md`
- Dashboard: `.docs/MONITORING_DASHBOARD.md`
