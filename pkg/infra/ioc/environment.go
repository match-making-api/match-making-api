package ioc

import (
	"os"

	"github.com/leet-gaming/match-making-api/pkg/infra/config"
)

// EnvironmentConfig retrieves configuration settings from environment variables.
//
// Returns:
//   - config.Config: A struct containing the populated configuration settings.
//   - error: An error if any issues occur during the configuration process (currently always nil).
func EnvironmentConfig() (config.Config, error) {
	config := config.Config{
		Auth: config.AuthConfig{
			SteamConfig: config.SteamConfig{
				SteamKey:    os.Getenv("STEAM_KEY"),
				PublicKey:   os.Getenv("STEAM_PUB_KEY"),
				Certificate: os.Getenv("STEAM_CERT"),
				VHashSource: os.Getenv("STEAM_VHASH_SOURCE"),
			},
		},
		MongoDB: config.MongoDBConfig{
			URI:         os.Getenv("MONGO_URI"),
			PublicKey:   os.Getenv("MONGO_PUB_KEY"),
			Certificate: os.Getenv("MONGO_CERT"),
			DBName:      os.Getenv("MONGO_DB_NAME"),
		},
		Kafka: config.KafkaConfig{
			Brokers:            os.Getenv("KAFKA_BOOTSTRAP"),
			Version:            os.Getenv("KAFKA_VERSION"),
			Group:              os.Getenv("KAFKA_CONSUMER_GROUP_ID"),
			AssignmentStrategy: os.Getenv("KAFKA_PARTITION_ASSIGNMENT_STRATEGY"),
		},
		Api: config.ApiConfig{
			RID:           os.Getenv("RID_SERVICE_URL"),
			PlayerProfile: os.Getenv("PLAYER_PROFILE_SERVICE_URL"),
			Subscription:  os.Getenv("SUBSCRIPTION_SERVICE_URL"),
		},
	}

	return config, nil
}
