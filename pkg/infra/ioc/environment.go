package ioc

import (
	"os"

	common "github.com/leet-gaming/match-making-api/pkg/domain"
)

// EnvironmentConfig retrieves configuration settings from environment variables.
//
// Returns:
//   - common.Config: A struct containing the populated configuration settings.
//   - error: An error if any issues occur during the configuration process (currently always nil).
func EnvironmentConfig() (common.Config, error) {
	config := common.Config{
		Auth: common.AuthConfig{
			SteamConfig: common.SteamConfig{
				SteamKey:    os.Getenv("STEAM_KEY"),
				PublicKey:   os.Getenv("STEAM_PUB_KEY"),
				Certificate: os.Getenv("STEAM_CERT"),
				VHashSource: os.Getenv("STEAM_VHASH_SOURCE"),
			},
		},
		MongoDB: common.MongoDBConfig{
			URI:         os.Getenv("MONGO_URI"),
			PublicKey:   os.Getenv("MONGO_PUB_KEY"),
			Certificate: os.Getenv("MONGO_CERT"),
			DBName:      os.Getenv("MONGO_DB_NAME"),
		},
	}

	return config, nil
}
