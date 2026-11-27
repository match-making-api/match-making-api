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
	"github.com/leet-gaming/match-making-api/pkg/infra/traffic"
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

	envCfg, _ := ioc.EnvironmentConfig()
	traffic.SyncConfiguredPercentGauge(envCfg.TrafficMigration)
	slog.Info("Starting PlayerQueuedConsumer for matchmaking.commands topic",
		"new_path_enabled", envCfg.DualWrite.Enabled,
		"new_path_failure_policy", envCfg.DualWrite.FailurePolicy,
		"new_path_only", envCfg.DualWrite.IsNewPathOnly(),
		"new_path_operations", envCfg.DualWrite.Operations(),
		"traffic_new_path_percent", envCfg.TrafficMigration.NewPathPercent,
		"traffic_strategy", envCfg.TrafficMigration.Strategy,
		"traffic_region_allowlist", envCfg.TrafficMigration.RegionAllowlist,
		"legacy_routing_retired", envCfg.TrafficMigration.IsLegacyRoutingRetired(),
		"consistency_check_enabled", envCfg.Consistency.Enabled)
	if err := consumer.Start(ctx); err != nil {
		slog.Error("PlayerQueuedConsumer stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("Matchmaking commands consumer shut down gracefully")
}
