package infra

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// InjectKafka sets up Kafka client and event publisher in the container.
//
// Parameters:
//   - c: A container.Container instance used for dependency injection.
//
// Returns:
//   - error: An error if the injection process fails, nil otherwise.
func InjectKafka(c container.Container) error {
	// Kafka Client
	err := c.Singleton(func() (*kafka.Client, error) {
		config := kafka.NewConfigFromEnv()
		client, err := kafka.NewClient(config)
		if err != nil {
			return nil, err
		}
		return client, nil
	})

	if err != nil {
		return err
	}

	// Event Publisher
	err = c.Singleton(func() (*kafka.EventPublisher, error) {
		var client *kafka.Client
		if err := c.Resolve(&client); err != nil {
			return nil, err
		}
		return kafka.NewEventPublisher(client), nil
	})

	if err != nil {
		return err
	}

	return nil
}