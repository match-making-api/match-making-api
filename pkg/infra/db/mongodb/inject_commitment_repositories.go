package mongodb

import (
	"log/slog"

	"github.com/golobby/container/v3"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// InjectCommitmentRepositories registers commitment, game connection info, and push token
// repositories as singletons in the container.
func InjectCommitmentRepositories(c container.Container) error {
	// CommitmentRepository implements both CommitmentWriter and CommitmentReader
	err := c.Singleton(func(client *mongo.Client, cfg config.Config) (*CommitmentRepository, error) {
		return NewCommitmentRepository(client, cfg.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to register CommitmentRepository")
		return err
	}

	err = c.Singleton(func(repo *CommitmentRepository) (pairing_out.CommitmentWriter, error) {
		return repo, nil
	})
	if err != nil {
		slog.Error("Failed to register CommitmentWriter")
		return err
	}

	err = c.Singleton(func(repo *CommitmentRepository) (pairing_out.CommitmentReader, error) {
		return repo, nil
	})
	if err != nil {
		slog.Error("Failed to register CommitmentReader")
		return err
	}

	// GameConnectionInfoRepository implements both Writer and Reader
	err = c.Singleton(func(client *mongo.Client, cfg config.Config) (*GameConnectionInfoRepository, error) {
		return NewGameConnectionInfoRepository(client, cfg.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to register GameConnectionInfoRepository")
		return err
	}

	err = c.Singleton(func(repo *GameConnectionInfoRepository) (pairing_out.GameConnectionInfoWriter, error) {
		return repo, nil
	})
	if err != nil {
		slog.Error("Failed to register GameConnectionInfoWriter")
		return err
	}

	err = c.Singleton(func(repo *GameConnectionInfoRepository) (pairing_out.GameConnectionInfoReader, error) {
		return repo, nil
	})
	if err != nil {
		slog.Error("Failed to register GameConnectionInfoReader")
		return err
	}

	// PushTokenRepository implements both Writer and Reader
	err = c.Singleton(func(client *mongo.Client, cfg config.Config) (*PushTokenRepository, error) {
		return NewPushTokenRepository(client, cfg.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to register PushTokenRepository")
		return err
	}

	err = c.Singleton(func(repo *PushTokenRepository) (pairing_out.PushTokenWriter, error) {
		return repo, nil
	})
	if err != nil {
		slog.Error("Failed to register PushTokenWriter")
		return err
	}

	err = c.Singleton(func(repo *PushTokenRepository) (pairing_out.PushTokenReader, error) {
		return repo, nil
	})
	if err != nil {
		slog.Error("Failed to register PushTokenReader")
		return err
	}

	return nil
}
