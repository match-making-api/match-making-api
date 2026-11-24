package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/leet-gaming/match-making-api/pkg/domain"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	"github.com/leet-gaming/match-making-api/pkg/infra"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc"
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
		slog.Info("Received shutdown signal, stopping consistency worker...", "signal", sig)
		cancel()
	}()

	var ticker *usecases.ConsistencyTicker
	if err := c.Resolve(&ticker); err != nil {
		slog.Error("Failed to resolve ConsistencyTicker", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting ConsistencyTicker for legacy vs new-path monitoring (2509-002)")
	if err := ticker.Start(ctx); err != nil {
		slog.Error("ConsistencyTicker stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("Consistency check worker shut down gracefully")
}
