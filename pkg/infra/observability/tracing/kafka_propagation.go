package tracing

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// KafkaHeaderCarrier adapts kafka-go headers for OTel propagation.
type KafkaHeaderCarrier []kafka.Header

func (c KafkaHeaderCarrier) Get(key string) string {
	for _, h := range c {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *KafkaHeaderCarrier) Set(key, value string) {
	*c = append(*c, kafka.Header{Key: key, Value: []byte(value)})
}

func (c KafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(c))
	for i, h := range c {
		keys[i] = h.Key
	}
	return keys
}

// ExtractContext reads W3C trace context and correlation ID from Kafka headers.
func ExtractContext(ctx context.Context, headers []kafka.Header) (context.Context, string) {
	carrier := KafkaHeaderCarrier(headers)
	ctx = otel.GetTextMapPropagator().Extract(ctx, &carrier)
	return ctx, carrier.Get(HeaderCorrelationID)
}

// InjectContext writes trace context and correlation ID into header map.
func InjectContext(ctx context.Context, headers map[string]string, correlationID string) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}
	carrier := propagation.MapCarrier(headers)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if correlationID != "" {
		headers[HeaderCorrelationID] = correlationID
	}
	return headers
}

// InjectContextToKafkaHeaders writes trace context into kafka.Header slice.
func InjectContextToKafkaHeaders(ctx context.Context, headers []kafka.Header, correlationID string) []kafka.Header {
	m := make(map[string]string, len(headers)+2)
	for _, h := range headers {
		m[h.Key] = string(h.Value)
	}
	m = InjectContext(ctx, m, correlationID)
	out := make([]kafka.Header, 0, len(m))
	for k, v := range m {
		out = append(out, kafka.Header{Key: k, Value: []byte(v)})
	}
	return out
}
