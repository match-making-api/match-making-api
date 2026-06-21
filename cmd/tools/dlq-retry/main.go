// Command dlq-retry re-publishes selected messages from matchmaking.dlq to their original topic.
//
// Usage:
//
//	go run ./cmd/tools/dlq-retry -brokers=localhost:9092 -limit=10 -dry-run
//
// Safety: consumers must be idempotent. Always dry-run first. Ops ownership: platform team.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

func main() {
	brokers := flag.String("brokers", "localhost:9092", "Kafka bootstrap servers")
	limit := flag.Int("limit", 10, "Max DLQ messages to retry")
	dryRun := flag.Bool("dry-run", true, "Log actions without publishing")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{*brokers},
		Topic:   kafka.TopicDLQ,
		GroupID: fmt.Sprintf("dlq-retry-tool-%d", time.Now().Unix()),
		MaxWait: time.Second,
	})
	defer reader.Close()

	client, err := kafka.NewClient(&kafka.Config{BootstrapServers: *brokers})
	if err != nil {
		slog.Error("Failed to create Kafka client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	retried := 0
	for retried < *limit {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			slog.Error("Fetch failed", "error", err)
			continue
		}

		var dlq kafka.DLQMessage
		if err := json.Unmarshal(msg.Value, &dlq); err != nil {
			slog.Warn("Skipping non-DLQ payload", "offset", msg.Offset, "error", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		slog.Info("DLQ message candidate",
			"original_topic", dlq.OriginalTopic,
			"original_offset", dlq.OriginalOffset,
			"retry_count", dlq.RetryCount,
			"error", dlq.Error,
			"dry_run", *dryRun)

		if !*dryRun {
			key := dlq.OriginalKey
			if key == "" {
				key = fmt.Sprintf("dlq-retry-%d", msg.Offset)
			}
			if err := client.PublishBytes(ctx, dlq.OriginalTopic, key, dlq.Payload, map[string]string{
				"dlq_retry":    "true",
				"retry_count":  fmt.Sprintf("%d", dlq.RetryCount+1),
			}); err != nil {
				slog.Error("Republish failed", "error", err)
				continue
			}
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("Commit failed", "error", err)
			continue
		}
		retried++
	}

	slog.Info("DLQ retry complete", "processed", retried, "dry_run", *dryRun)
}
