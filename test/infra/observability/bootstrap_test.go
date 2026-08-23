package observability_test

import (
	"testing"

	"github.com/leet-gaming/match-making-api/pkg/infra/observability"
	"github.com/stretchr/testify/require"
)

func TestInit_DisabledTracing(t *testing.T) {
	t.Setenv("OTEL_ENABLED", "false")
	t.Setenv("METRICS_ENABLED", "false")

	require.NoError(t, observability.Init())
}
