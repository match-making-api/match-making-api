# Queue subscription check (2506-001)

Match-making validates subscription **tier** and **queue-join usage** before adding a player to the pool. It does **not** call billing execution RPCs (`ValidateOperation`, `ConfirmOperation`), deduct balances, or update usage. Wallet API owns those writes.

## When

`HandlePlayerQueuedProto` runs the check after the player id is parsed and before region lookup or pool insert.

| Snapshot | Priority boost | Result |
|----------|----------------|--------|
| Missing reader, or nil snapshot | no | Basic queue allowed |
| Missing reader, or nil snapshot | yes | `upgrade required` |
| Plan free / basic, status active | yes | `upgrade required` |
| Plan premium / priority / plus / pro | yes | Allowed (2506-002 may then publish the chargeable event) |
| `available.queue_join` <= 0 | either | `usage limit exceeded` |
| Status cancelled / expired / other non-active | either | `subscription inactive` |
| GetSubscription `is_valid=false` | either | `subscription inactive` |

RPC errors fail closed: the player is not added to the pool.

## Wiring

`QueueSubscriptionLookup` calls `SubscriptionService.GetSubscription` only when `config.Api.Subscription` is set. Without that address, premium options stay gated and the basic queue still proceeds.

Usage key read from the subscription map: `queue_join`.
