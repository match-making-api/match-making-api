package observability

import (
	"context"

	"github.com/leet-gaming/match-making-api/pkg/infra/observability/metrics"
	"github.com/leet-gaming/match-making-api/pkg/infra/observability/tracing"
)

// Init configures OpenTelemetry tracing and the Prometheus metrics server.
// Safe to call from REST API and worker processes.
func Init() error {
	if shutdown, err := tracing.Init(context.Background()); err != nil {
		return err
	} else if shutdown != nil {
		// Shutdown runs on process exit via defer in main if needed; no-op hook here.
		_ = shutdown
	}

	metrics.StartServerIfEnabled()

	return nil
}
