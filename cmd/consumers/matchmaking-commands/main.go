package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/leet-gaming/match-making-api/pkg/domain"
	"github.com/leet-gaming/match-making-api/pkg/infra"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Slim inject — only Kafka, Redis, MongoDB (regions/games), and pairing domain.
	// Skips IAM, billing, squad, lobbies which are REST API-only concerns.
	builder := ioc.NewContainerBuilder()
	c := builder.WithEnvFile().With(infra.InjectWorker).With(domain.InjectWorker).Build()
	defer builder.Close(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		slog.Info("Received shutdown signal, stopping consumer...", "signal", sig)
		cancel()
	}()

	// Resolve and start the PlayerQueuedConsumer
	var consumer *kafka.PlayerQueuedConsumer
	if err := c.Resolve(&consumer); err != nil {
		slog.Error("Failed to resolve PlayerQueuedConsumer", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting PlayerQueuedConsumer for matchmaking.commands topic")
	if err := consumer.Start(ctx); err != nil {
		slog.Error("PlayerQueuedConsumer stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("Matchmaking commands consumer shut down gracefully")
}
