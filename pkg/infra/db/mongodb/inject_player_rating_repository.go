package mongodb

import (
	"log/slog"

	"github.com/golobby/container/v3"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// InjectPlayerRatingRepository registers PlayerRatingRepository as a singleton.
func InjectPlayerRatingRepository(c container.Container) error {
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) pairing_out.PlayerRatingRepository {
		return NewPlayerRatingRepository(client, cfg.MongoDB.DBName)
	})
	if err != nil {
		slog.Error("Failed to register PlayerRatingRepository")
		return err
	}
	return nil
}
