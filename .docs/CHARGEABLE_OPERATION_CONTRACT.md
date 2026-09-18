# Chargeable Operation Contract (Wallet API)

Refs **2506-002**. Match-making-api **publishes** `ChargeableOperationRequested` when a paid option applies. **Wallet API** consumes the event and executes billing, balance deduction, and subscription usage. Match-making never calls billing RPCs or updates wallets.

## Topic

| Constant | Topic |
|----------|--------|
| `TopicChargeableRequested` | `matchmaking.billing.chargeable` |

Direction: **match-making-api → wallet-api**. Partition key: `player_id`.

## When published (queue join)

| Condition | Publish? |
|-----------|----------|
| `PlayerQueued.priority_boost` absent or `<= 0` | **No** |
| `priority_boost > 0` after successful pool add | **Yes** (`operation_type=queue_priority_boost`) |

Amount: `PRIORITY_BOOST_AMOUNT_CENTS` env (default **100**). If `priority_boost > 1`, that value is used as `amount_cents` (producer-specified fee).

## Event JSON (CloudEvents-shaped)

```json
{
  "id": "...",
  "type": "ChargeableOperationRequested",
  "source": "match-making-api",
  "specversion": "1.0",
  "time_unix_ms": 0,
  "subject": "<player_id>",
  "resource_owner_id": "...",
  "correlation_id": "...",
  "dataschema_version": 1,
  "chargeable_operation_requested": {
    "operation_type": "queue_priority_boost",
    "amount_cents": 100,
    "currency": "USD",
    "player_id": "...",
    "resource_owner_id": "...",
    "tenant_id": "...",
    "client_id": "...",
    "game_id": "...",
    "region": "...",
    "correlation_id": "...",
    "idempotency_key": "queue_priority_boost:<player_id>:<game_id>:<correlation_id>",
    "requested_at_epoch_ms": 0
  }
}
```

Headers: `ce_type=ChargeableOperationRequested`, `ce_source=match-making-api`.

## Idempotency

Wallet MUST treat `idempotency_key` as unique. Retries of the same PlayerQueued (same correlation) must not double-charge.

## Proto

Message `ChargeableOperationRequestedPayload` is documented in `matchmaking_events.proto`. Until codegen adds it to the `MatchmakingEvent` oneof, producers use the hand-written JSON type in `chargeable_operation.go`.

## Code

- Publish: `EventPublisher.PublishChargeableOperationRequested`
- Trigger: `MatchmakingEventConsumer.publishChargeableIfPriorityBoost`
- Future: tournament entry (2506-003), lobby creation fee (2507-001) reuse the same topic/payload shape
