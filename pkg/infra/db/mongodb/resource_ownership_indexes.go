package mongodb

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ResourceOwnershipIndexModels returns compound indexes for nested resource_owner
// (Refs 2508-002). Safe to CreateMany on collections that use common.BaseEntity
// or dual-write nested ownership (lobby, player_ratings, match_results).
func ResourceOwnershipIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "resource_owner.tenant_id", Value: 1},
				{Key: "resource_owner.client_id", Value: 1},
			},
			Options: options.Index().SetName("idx_ro_tenant_client"),
		},
		{
			Keys: bson.D{
				{Key: "resource_owner.tenant_id", Value: 1},
				{Key: "resource_owner.user_id", Value: 1},
			},
			Options: options.Index().SetName("idx_ro_tenant_user"),
		},
		{
			Keys: bson.D{
				{Key: "resource_owner.tenant_id", Value: 1},
				{Key: "resource_owner.group_id", Value: 1},
			},
			Options: options.Index().SetName("idx_ro_tenant_group"),
		},
	}
}

// FlatTenantClientIndexModels indexes legacy top-level tenant_id + client_id
// during dual-write migration (2508-001 → 2508-002).
func FlatTenantClientIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "client_id", Value: 1},
			},
			Options: options.Index().SetName("idx_flat_tenant_client"),
		},
	}
}

// EnsureResourceOwnershipIndexes creates ownership indexes on collection.
// Errors are logged and do not fail repository construction.
func EnsureResourceOwnershipIndexes(ctx context.Context, collection *mongo.Collection, includeFlat bool) {
	if collection == nil {
		return
	}
	indexes := ResourceOwnershipIndexModels()
	if includeFlat {
		indexes = append(indexes, FlatTenantClientIndexModels()...)
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Warn("Failed to create resource ownership indexes",
			"collection", collection.Name(),
			"error", err)
		return
	}
	slog.Info("Resource ownership indexes ensured", "collection", collection.Name())
}
