# Runbook: Dual-write rollback (2509-001)

**Trigger:** Dual-write divergence, match-making new path errors, or planned rollback before traffic migration (2509-003)  
**Severity:** High when players stuck / matches not forming on new path  
**Owner:** Matchmaking platform (`#matchmaking-ops`) + Replay API (`#replay-api`)

## Goal

Disable the **new** Kafka / match-making-api path while keeping the **legacy** replay-api matchmaking session path.

## Immediate rollback (new path off)

1. **Stop producing** (replay-api): set `KAFKA_MATCHMAKING_PRODUCER_ENABLED=false` and redeploy / reload config.
2. **Stop consuming / processing** (match-making-api): set `MATCHMAKING_DUAL_WRITE_ENABLED=false` on `matchmaking-commands` consumer and restart.
3. Confirm join/leave succeed via legacy session APIs only.
4. Confirm no new `PlayerQueued` / `MatchCreated` traffic on dashboards (2505-001).

## Checks

1. Dashboard produce/consume rates on `matchmaking.commands` and `matchmaking.matches.created`.
2. Consumer logs: `Dual-write new path disabled; skipping PlayerQueued|PlayerLeftQueue`.
3. replay-api join still saves sessions; produce skipped when producer disabled.
4. Pool depth should stop growing; drain residual pool entries if needed (ops judgment).

## Re-enable dual-write

1. Set `KAFKA_MATCHMAKING_PRODUCER_ENABLED=true` (replay-api).
2. Set `MATCHMAKING_DUAL_WRITE_ENABLED=true` (match-making-api) and restart consumer.
3. Spot-check join → pool insert → match create → `MatchCreated`.
4. Hand off consistency monitoring to 2509-002 before increasing traffic (2509-003).

## Failure policy note

`MATCHMAKING_DUAL_WRITE_FAILURE_POLICY=legacy_only_and_alert` (default): MatchCreated produce failures keep the local pair and log errors — do not treat as full request failure unless policy is switched to `fail_request`.

## Links

- Dual-write doc: `.docs/DUAL_WRITE.md`
- Lag: [consumer-lag-spike.md](./consumer-lag-spike.md)
- Monitoring: `.docs/MONITORING_DASHBOARD.md`
