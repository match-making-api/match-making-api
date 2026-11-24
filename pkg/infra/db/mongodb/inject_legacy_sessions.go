package mongodb

import (
	"log/slog"

	"github.com/golobby/container/v3"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// InjectLegacySessionsRepository registers LegacySessionReader for consistency monitoring.
func InjectLegacySessionsRepository(c container.Container) error {
	return c.Singleton(func(client *mongo.Client, cfg config.Config) pairing_out.LegacySessionReader {
		dbName := cfg.Consistency.LegacyDBName
		if dbName == "" {
			dbName = cfg.MongoDB.DBName
		}
		collection := cfg.Consistency.LegacyCollection
		if collection == "" {
			collection = DefaultLegacySessionsCollection
		}
		slog.Info("Registered LegacySessionReader",
			"db", dbName,
			"collection", collection)
		return NewLegacySessionsRepository(client, dbName, collection)
	})
}
