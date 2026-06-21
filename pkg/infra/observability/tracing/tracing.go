package tracing

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	HeaderTraceParent   = "traceparent"
	HeaderTraceState    = "tracestate"
	HeaderCorrelationID = "x-correlation-id"
)

var tracer = otel.Tracer("match-making-api/kafka")

// Init configures the global OpenTelemetry tracer provider.
// Set OTEL_EXPORTER_OTLP_ENDPOINT (default localhost:4317) and OTEL_SERVICE_NAME.
func Init(ctx context.Context) (func(context.Context) error, error) {
	if os.Getenv("OTEL_ENABLED") == "false" {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		return func(context.Context) error { return nil }, nil
	}

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317"
	}
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "match-making-api"
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		// Degrade gracefully when collector is unavailable (local dev).
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// EnsureCorrelationID returns an existing correlation ID or generates one.
func EnsureCorrelationID(existing string) string {
	if existing != "" {
		return existing
	}
	return uuid.New().String()
}

// StartKafkaConsumeSpan starts a consumer span with correlation attributes.
func StartKafkaConsumeSpan(ctx context.Context, topic, groupID, correlationID string) (context.Context, trace.Span) {
	return tracer.Start(ctx, "kafka.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", topic),
			attribute.String("messaging.kafka.consumer.group", groupID),
			attribute.String("correlation_id", correlationID),
		),
	)
}

// StartKafkaProduceSpan starts a producer span.
func StartKafkaProduceSpan(ctx context.Context, topic, correlationID string) (context.Context, trace.Span) {
	return tracer.Start(ctx, "kafka.produce",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", topic),
			attribute.String("correlation_id", correlationID),
		),
	)
}
