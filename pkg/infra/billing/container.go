package billing

import (
	"log/slog"

	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Inject initializes and sets up the billing module within the given container.
//
// Parameters:
//   - container: A container.Container that serves as a dependency injection container
//     for the billing module. It may contain configurations or dependencies
//     required for the module's initialization.
//
// Returns:
//   - An error if the injection process encounters any issues, or nil if successful.
func Inject(c container.Container) error {
	c.Singleton(func(config config.Config) (SubscriptionServiceClient, error) {
		serverAddress := config.Api.Subscription

		conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			slog.Error("Failed to connect to PlayerProfile service", "err", err)
			return nil, err
		}

		return NewSubscriptionServiceClient(conn), nil

	})

	return nil
}
