package mongodb

import (
	"log/slog"

	"github.com/golobby/container/v3"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// InjectMatchResultRepository registers MatchResultRepository as a singleton.
func InjectMatchResultRepository(c container.Container) error {
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) pairing_out.MatchResultRepository {
		return NewMatchResultRepository(client, cfg.MongoDB.DBName)
	})
	if err != nil {
		slog.Error("Failed to register MatchResultRepository")
		return err
	}
	return nil
}
