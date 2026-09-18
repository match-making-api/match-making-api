# Resource Ownership Migration — Refs 2508-001

**Scope:** match-making-api MongoDB collections only.  
**replay-api:** out of this PR — keep story `partial` until handoff.

## Goal

Canonical nested field `resource_owner: { tenant_id, client_id, group_id, user_id }` on persisted matchmaking documents, while **keeping flat** `tenant_id` / `client_id` / `creator_id` / `resource_owner_id` for dual-write and rollback.

## Already on BaseEntity

Entities embedding `common.BaseEntity` already have nested `resource_owner`: Pair, Invitation, Notification, Game, GameMode, Region, Commitment, etc.

## Migrated in this story (MM)

| Collection | Flat source | Nested target |
|------------|-------------|-----------------|
| `lobbies` | `tenant_id`, `client_id`, `creator_id` → user | `resource_owner.*` |
| `player_ratings` | `tenant_id`, `client_id`, `resource_owner_id` | `resource_owner.*` |
| `match_results` | same | `resource_owner.*` |

Application code: `EnsureResourceOwner` / `ValidateOwnership` on write; helpers in `pkg/common/resource_owner_helpers.go`.

## Backfill (mongosh)

```bash
mongosh "$MONGO_URI" --file deploy/migrations/2508-001-backfill-resource-owner.js
```

Script is **idempotent** (`$or` missing nested fields). Dry-run: set `DRY_RUN = true` at top of the file.

## Validation query

```javascript
db.lobbies.countDocuments({
  $or: [
    { "resource_owner.tenant_id": { $exists: false } },
    { "resource_owner.client_id": { $exists: false } },
  ]
})
// expect 0 after successful backfill (for docs that have flat tenant_id)
```

## Rollback

1. **Do not drop flat fields** during this phase.
2. To undo nested field only:

```javascript
db.lobbies.updateMany({}, { $unset: { resource_owner: "" } })
db.player_ratings.updateMany({}, { $unset: { resource_owner: "" } })
db.match_results.updateMany({}, { $unset: { resource_owner: "" } })
```

3. Redeploy previous app version that ignores `resource_owner`.

## Indexes (2508-002)

Compound indexes on `resource_owner.*` (and flat dual-write keys): see [.docs/RESOURCE_OWNERSHIP_INDEXES.md](RESOURCE_OWNERSHIP_INDEXES.md).

## Not migrated (intentionally)

| Entity | Reason |
|--------|--------|
| `Pool` | In-memory / ephemeral |
| `Party` / `Peer` / `Match` stubs | Minimal IDs only; ownership flows via Pair / events |
| `Schedule` | Not persisted with ownership yet |
