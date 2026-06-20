# Error Handling and DLQ (2505-003)

Dead-letter queue flow for match-making-api Kafka consumers.

## Topic

| Name | Constant | Notes |
|------|----------|-------|
| `matchmaking.dlq` | `TopicDLQ` | Equivalent to epic `matchmaking.dead-letter` |

Create via Strimzi KafkaTopic CR in your cluster overlay.

## Retry policy

| Env | Default | Description |
|-----|---------|-------------|
| `KAFKA_MAX_RETRIES` | `3` | Processing attempts before DLQ |
| `KAFKA_RETRY_BACKOFF_MS` | `500` | Linear backoff multiplier per attempt |

## DLQ message format

```json
{
  "original_topic": "matchmaking.commands",
  "original_partition": 0,
  "original_offset": 12345,
  "original_key": "player-uuid",
  "error": "processing failed: ...",
  "retry_count": 3,
  "timestamp": 1710000000000,
  "payload": "<original bytes>"
}
```

## Commit semantics

- **Success:** offset committed.
- **Retry exhausted + DLQ publish OK:** offset committed (message removed from main topic processing; contained in DLQ).
- **DLQ publish fails:** offset **not** committed; message will be retried.

## Manual retry

Tool: `cmd/tools/dlq-retry`

```bash
# Inspect only
go run ./cmd/tools/dlq-retry -brokers=localhost:9092 -limit=5 -dry-run=true

# Republish to original topics (requires idempotent handlers)
go run ./cmd/tools/dlq-retry -brokers=localhost:9092 -limit=5 -dry-run=false
```

**Ownership:** Platform/on-call only. Coordinate with matchmaking team before bulk retry.

## Monitoring

Alert when DLQ depth > 0 (Kafka Exporter lag on `matchmaking.dlq` or custom metric). See 2505-001 dashboard panels and `.docs/runbooks/dlq-filling-up.md`.

## Idempotency

Consumers must tolerate redelivery. Use idempotency keys / processed stores (see ratings, prizes, analytics handlers).
