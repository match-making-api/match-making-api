package mongodb

import (
	"log/slog"

	"github.com/golobby/container/v3"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// InjectRatingsProcessedStore registers RatingsProcessedStore as a singleton.
func InjectRatingsProcessedStore(c container.Container) error {
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) pairing_out.RatingsProcessedStore {
		return NewRatingsProcessedStore(client, cfg.MongoDB.DBName)
	})
	if err != nil {
		slog.Error("Failed to register RatingsProcessedStore")
		return err
	}
	return nil
}
