package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestInjectExtractRoundTrip(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	ctx, span := otel.Tracer("test").Start(context.Background(), "parent")
	defer span.End()

	headers := InjectContextToKafkaHeaders(ctx, nil, "corr-123")
	extracted, corrID := ExtractContext(context.Background(), headers)

	assert.Equal(t, "corr-123", corrID)
	assert.NotEqual(t, context.Background(), extracted)
}

func TestEnsureCorrelationID(t *testing.T) {
	assert.Equal(t, "existing", EnsureCorrelationID("existing"))
	generated := EnsureCorrelationID("")
	require.NotEmpty(t, generated)
}

func TestKafkaHeaderCarrier(t *testing.T) {
	var c KafkaHeaderCarrier
	c.Set("traceparent", "00-abc-def-01")
	assert.Equal(t, "00-abc-def-01", c.Get("traceparent"))
	assert.Contains(t, c.Keys(), "traceparent")
}
