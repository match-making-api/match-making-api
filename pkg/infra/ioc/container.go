package ioc

import (
	"context"
	"log/slog"
	"os"

	container "github.com/golobby/container/v3"
	"github.com/joho/godotenv"
	common "github.com/leet-gaming/match-making-api/pkg/domain"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

	err := b.Container.Singleton(func() (common.Config, error) {
		return EnvironmentConfig()
	})

	if err != nil {
		slog.Error("Failed to load EnvironmentConfig.")
		panic(err)
	}

	return b
}

// WithInboundPorts configures the ContainerBuilder with inbound ports for various services.
// It sets up singleton instances for different commands and readers used in the application.
//
// Parameters:
//   - b: A pointer to the ContainerBuilder instance being configured.
//
// Returns:
//   - *ContainerBuilder: The same ContainerBuilder instance, allowing for method chaining.
func (b *ContainerBuilder) WithInboundPorts() *ContainerBuilder {
	return b
}

// InjectMongoDB registers a MongoDB client as a singleton in the provided container.
// It configures the MongoDB client using the application's configuration and establishes
// a connection to the MongoDB server.
//
// Parameters:
//   - c: A container.Container instance where the MongoDB client will be registered.
//
// Returns:
//   - error: An error if the MongoDB client registration or connection fails, nil otherwise.
func InjectMongoDB(c container.Container) error {
	err := c.Singleton(func() (*mongo.Client, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for mongo.Client.", "err", err)
			return nil, err
		}

		mongoOptions := options.Client().ApplyURI(config.MongoDB.URI).SetRegistry(db.MongoRegistry).SetMaxPoolSize(100)

		client, err := mongo.Connect(context.TODO(), mongoOptions)

		if err != nil {
			slog.Error("Failed to connect to MongoDB.", "err", err)
			return nil, err
		}

		return client, nil
	})

	if err != nil {
		slog.Error("Failed to load mongo.Client.")
		return err
	}

	return nil
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
