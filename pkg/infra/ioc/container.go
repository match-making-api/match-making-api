package ioc

import (
	"context"
	"log/slog"
	"os"

	container "github.com/golobby/container/v3"
	"github.com/joho/godotenv"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc/db/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

// ContainerBuilder is a container builder for the application.
type ContainerBuilder struct {
	Container container.Container
}

// NewContainerBuilder creates and initializes a new ContainerBuilder instance.
//
// Panics:
//   - error: If any error occurs during the container setup.
//
// Returns:
//   - *ContainerBuilder: A pointer to the newly created and initialized ContainerBuilder.
func NewContainerBuilder() *ContainerBuilder {
	c := container.New()

	b := &ContainerBuilder{
		c,
	}

	err := c.Singleton(func() container.Container {
		return b.Container
	})

	if err != nil {
		slog.Error("Failed to register *container.Container  in NewContainerBuilder.")
		panic(err)
	}

	err = c.Singleton(func() *ContainerBuilder {
		return b
	})

	if err != nil {
		slog.Error("Failed to register *ContainerBuilder in NewContainerBuilder.")
		panic(err)
	}

	return b
}

// Build returns the configured Container instance from the ContainerBuilder.
//
// Returns:
//   - container.Container: The fully configured Container instance.
func (b *ContainerBuilder) Build() container.Container {
	return b.Container
}

// WithEnvFile configures the ContainerBuilder to load environment variables from a .env file.
//
// Parameters:
//   - b: A pointer to the ContainerBuilder instance being configured.
//
// Panics:
//   - err: An error if the.env file could not be loaded.
//
// Returns:
//   - *ContainerBuilder: The same ContainerBuilder instance, allowing for method chaining.
func (b *ContainerBuilder) WithEnvFile() *ContainerBuilder {
	if os.Getenv("DEV_ENV") == "true" || os.Getenv("DEV_ENV") == "" {
		err := godotenv.Load()
		if err != nil {
			slog.Error("Failed to load .env file")
			panic(err)
		}
	}

	err := b.Container.Singleton(func() (config.Config, error) {
		return EnvironmentConfig()
	})

	if err != nil {
		slog.Error("Failed to load EnvironmentConfig.")
		panic(err)
	}

	return b
}

// With registers a resolver as a singleton in the container.
// It's used to add custom dependencies to the container.
//
// Parameters:
//   - resolver: An interface{} representing the resolver to be registered as a singleton.
//     This can be any type that implements the necessary resolution logic.
//
// Returns:
//   - *ContainerBuilder: Returns the ContainerBuilder instance to allow for method chaining.
//
// Panics:
//   - If there's an error registering the resolver, this method will log the error and panic.
func (b *ContainerBuilder) With(resolver interface{}) *ContainerBuilder {
	c := b.Container

	err := c.Singleton(resolver)

	if err != nil {
		slog.Error("Failed to register resolver.", "err", err)
		panic(err)
	}

	return b
}

// Close attempts to close the MongoDB client connection associated with the container.
//
// Parameters:
//   - c: A container.Container instance that potentially contains a MongoDB client.
func (b *ContainerBuilder) Close(c container.Container) {
	var client *mongo.Client
	err := c.Resolve(&client)

	if client != nil && err == nil {
		client.Disconnect(context.TODO())
	}
}

func InjectIoc(c container.Container) error {
	return common.InjectAll(c, mongodb.InjectMongoDB)
}
