# Distributed Tracing (2505-002)

OpenTelemetry tracing for match-making-api Kafka producers and consumers.

## Enable

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_ENABLED` | `true` | Set `false` to disable export |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` | OTLP gRPC endpoint (Jaeger/Tempo collector) |
| `OTEL_SERVICE_NAME` | `match-making-api` | Service name in traces |

## Propagation

| Header | Purpose |
|--------|---------|
| `traceparent` | W3C Trace Context |
| `tracestate` | W3C optional vendor state |
| `x-correlation-id` | Matchmaking request correlation (also in `EventEnvelope.correlation_id`) |

Producers inject headers in `Client.PublishBytes`. Consumers extract and create child spans in `Consumer.processMessage`.

## Search traces

**Jaeger UI:** Search by `correlation_id` tag or `traceID` from logs.

**Tempo/Grafana:** `{ resource.service.name="match-making-api" && span.correlation_id="..." }`

## Spans

| Span | Kind | Attributes |
|------|------|------------|
| `kafka.produce` | Producer | topic, correlation_id |
| `kafka.consume` | Consumer | topic, group_id, correlation_id |

## replay-api handoff

See `specs/cross-repo/replay-api/replay-api-2505-002-distributed-tracing.md` in match-making-spec.

## Sampling

Default: all spans exported. For high volume, configure OTel sampler via `OTEL_TRACES_SAMPLER` env (see OpenTelemetry docs).
