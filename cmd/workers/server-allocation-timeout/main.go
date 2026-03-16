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
		slog.Info("Received shutdown signal, stopping worker...", "signal", sig)
		cancel()
	}()

	var worker *usecases.ServerAllocationTimeoutWorker
	if err := c.Resolve(&worker); err != nil {
		slog.Error("Failed to resolve ServerAllocationTimeoutWorker", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting ServerAllocationTimeoutWorker")
	if err := worker.Start(ctx); err != nil && err != context.Canceled {
		slog.Error("ServerAllocationTimeoutWorker stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("Server allocation timeout worker shut down gracefully")
}
