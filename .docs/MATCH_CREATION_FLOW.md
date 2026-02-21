# Match Creation Flow

End-to-end flow from PotentialMatchFound → validation → MatchCreated (Epic §10 Phase 2).

## Flow

1. **PotentialMatchFound** (implicit): `AddAndFindNextPair.Execute` returns a non-nil pair when `pool.Peek` dequeues enough compatible players.
2. **Validation**: `MatchValidator` runs before producing MatchCreated:
   - Resource ownership: `resource_owner_id` required
   - Required fields: `tenant_id`, `client_id` for downstream access control
   - Pair integrity: pair exists, has players, has match_id
3. **MatchCreated**: Published to `matchmaking.matches.created` with full Protobuf payload (Epic §9).

## Validation Rules

| Rule | Description |
|------|-------------|
| ResourceOwnershipRule | `resource_owner_id` must be present in envelope |
| RequiredFieldsRule | `tenant_id`, `client_id` must be present in payload |
| PairIntegrityRule | Pair must exist, have players, and valid match_id |

## Failure Modes

| Failure | Handling |
|---------|----------|
| Validation failure | Log error, skip MatchCreated. Pair remains in DB. No orphaned MatchCreated. |
| Produce failure | Log error, non-fatal. Match exists; replay-api may need reconciliation. |
| Duplicate PotentialMatchFound | Idempotency: partition by `match_id`; consumers dedupe. Pair creation is single-shot per Peek. |

## Idempotency

- MatchCreated keyed by `match_id` for Kafka partitioning.
- Consumers should dedupe by `match_id` when processing.
- Duplicate Peek for same parties is prevented by pool state (parties removed on Peek).
