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

	builder := ioc.NewContainerBuilder()
	c := builder.WithEnvFile().With(infra.InjectWorker).With(domain.InjectWorker).Build()
	defer builder.Close(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		slog.Info("Received shutdown signal, stopping consumer...", "signal", sig)
		cancel()
	}()

	var consumer *kafka.MatchStartedConsumer
	if err := c.Resolve(&consumer); err != nil {
		slog.Error("Failed to resolve MatchStartedConsumer", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting MatchStartedConsumer for matchmaking.match.started topic")
	if err := consumer.Start(ctx); err != nil {
		slog.Error("MatchStartedConsumer stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("Match started consumer shut down gracefully")
}
