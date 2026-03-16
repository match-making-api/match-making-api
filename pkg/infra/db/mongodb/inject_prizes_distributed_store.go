package mongodb

import (
	"log/slog"

	"github.com/golobby/container/v3"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// InjectPrizesDistributedStore registers PrizesDistributedStore as a singleton.
func InjectPrizesDistributedStore(c container.Container) error {
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) pairing_out.PrizesDistributedStore {
		return NewPrizesDistributedStore(client, cfg.MongoDB.DBName)
	})
	if err != nil {
		slog.Error("Failed to register PrizesDistributedStore")
		return err
	}
	return nil
}
