# Resource Ownership Indexes — Refs 2508-002

**Scope:** match-making-api only. replay-api indexes are a separate handoff.

## Strategy

| Index name | Keys | Collections |
|------------|------|-------------|
| `idx_ro_tenant_client` | `resource_owner.tenant_id`, `resource_owner.client_id` | lobbies, player_ratings, match_results, commitments, push_tokens, game_connection_info |
| `idx_ro_tenant_user` | `resource_owner.tenant_id`, `resource_owner.user_id` | same |
| `idx_ro_tenant_group` | `resource_owner.tenant_id`, `resource_owner.group_id` | same |
| `idx_flat_tenant_client` | `tenant_id`, `client_id` | lobbies, player_ratings, match_results (dual-write) |

Defined in `pkg/infra/db/mongodb/resource_ownership_indexes.go` and applied via `EnsureResourceOwnershipIndexes` when repositories start.

## Validation (app layer)

- Lobby create/update: `ValidateOwnership()` (tenant + client required) — Refs 2508-001
- PlayerRating / MatchResult Save: same
- Invalid ownership → repository returns error before write

## Performance smoke (manual / staging)

Epic target: ownership-filtered queries **&lt; 50ms** (p95). After deploy:

```javascript
// mongosh — confirm indexes
db.lobbies.getIndexes().filter(i => i.name && i.name.startsWith("idx_ro"))

// explain — should use idx_ro_tenant_client
db.lobbies.find({
  "resource_owner.tenant_id": UUID("…"),
  "resource_owner.client_id": UUID("…")
}).explain("executionStats")
```

Record `executionStats.executionTimeMillis` in the PR or runbook if &gt; 50ms on representative data.

## Out of scope

- Data backfill → 2508-001
- Session–Pool mapping → 2508-003
- replay-api collections → handoff
