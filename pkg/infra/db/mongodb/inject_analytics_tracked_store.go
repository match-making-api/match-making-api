package mongodb

import (
	"log/slog"

	"github.com/golobby/container/v3"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// InjectAnalyticsTrackedStore registers AnalyticsTrackedStore as a singleton.
func InjectAnalyticsTrackedStore(c container.Container) error {
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) pairing_out.AnalyticsTrackedStore {
		return NewAnalyticsTrackedStore(client, cfg.MongoDB.DBName)
	})
	if err != nil {
		slog.Error("Failed to register AnalyticsTrackedStore")
		return err
	}
	return nil
}
