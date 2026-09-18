// 2508-001 — Backfill nested resource_owner where flat fields are already UUIDs.
// Idempotent. Set DRY_RUN = true to only print counts.
//
// player_ratings / match_results keep string flat IDs; nested UUID resource_owner
// is written by the application on next Save (EnsureResourceOwner + dual-write).
// This script backfills lobbies only (UUID-typed flat fields).
//
// Usage:
//   mongosh "$MONGO_URI/$DB_NAME" --file deploy/migrations/2508-001-backfill-resource-owner.js

const DRY_RUN = false;

function backfillLobbies() {
  const filter = {
    $or: [
      { resource_owner: { $exists: false } },
      { "resource_owner.tenant_id": { $exists: false } },
      { "resource_owner.tenant_id": null },
    ],
    tenant_id: { $exists: true, $ne: null },
    client_id: { $exists: true, $ne: null },
  };
  const count = db.lobbies.countDocuments(filter);
  print(`lobbies to backfill: ${count}`);
  if (DRY_RUN || count === 0) return;
  const res = db.lobbies.updateMany(filter, [
    {
      $set: {
        resource_owner: {
          tenant_id: "$tenant_id",
          client_id: "$client_id",
          group_id: { $literal: null },
          user_id: "$creator_id",
        },
      },
    },
  ]);
  printjson(res);
}

function countRatingsMissingNested() {
  const n = db.player_ratings.countDocuments({
    resource_owner: { $exists: false },
    tenant_id: { $exists: true, $ne: "" },
  });
  print(`player_ratings without nested resource_owner (app dual-writes on Save): ${n}`);
}

function countMatchResultsMissingNested() {
  const n = db.match_results.countDocuments({
    resource_owner: { $exists: false },
    tenant_id: { $exists: true, $ne: "" },
  });
  print(`match_results without nested resource_owner (app dual-writes on Save): ${n}`);
}

print("=== 2508-001 resource_owner backfill ===");
backfillLobbies();
countRatingsMissingNested();
countMatchResultsMissingNested();
print("=== done ===");
