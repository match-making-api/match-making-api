package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/leet-gaming/match-making-api/cmd/rest-api/routing"
	"github.com/leet-gaming/match-making-api/pkg/domain"
	"github.com/leet-gaming/match-making-api/pkg/infra"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	builder := ioc.NewContainerBuilder()

	c := builder.WithEnvFile().With(infra.Inject).With(domain.Inject).Build()

	defer builder.Close(c)

	// Start Kafka consumer for matchmaking commands (PlayerQueued from replay-api)
	var playerQueuedConsumer *kafka.PlayerQueuedConsumer
	if err := c.Resolve(&playerQueuedConsumer); err != nil {
		slog.Warn("PlayerQueuedConsumer not available, skipping consumer startup", "err", err)
	} else {
		consumerCtx, consumerCancel := context.WithCancel(ctx)
		defer consumerCancel()

		go func() {
			slog.Info("Starting PlayerQueuedConsumer for matchmaking.commands topic")
			if err := playerQueuedConsumer.Start(consumerCtx); err != nil {
				slog.Error("PlayerQueuedConsumer stopped with error", "error", err)
			}
		}()
	}

	router := routing.NewRouter(ctx, c)

	slog.InfoContext(ctx, "Starting server on port 4991")

	http.ListenAndServe(":4991", router)
}
