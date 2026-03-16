package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/leet-gaming/match-making-api/cmd/rest-api/routing"
	"github.com/leet-gaming/match-making-api/pkg/domain"
	"github.com/leet-gaming/match-making-api/pkg/infra"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	builder := ioc.NewContainerBuilder()

	c := builder.WithEnvFile().With(infra.Inject).With(domain.Inject).Build()

	defer builder.Close(c)

	router := routing.NewRouter(ctx, c)

	slog.InfoContext(ctx, "Starting server on port 4991")

	server := &http.Server{
		Addr:           ":4991",
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB max header size
	}

	if err := server.ListenAndServe(); err != nil {
		slog.ErrorContext(ctx, "Server error", "err", err)
	}
}
