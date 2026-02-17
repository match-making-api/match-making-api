package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// Config holds Kafka client configuration.
//
// Security protocols:
//   - PLAINTEXT: no encryption, no auth (development only)
//   - SSL: TLS encryption, optional mTLS client certs
//   - SASL_PLAINTEXT: SASL auth over plaintext (not recommended)
//   - SASL_SSL: SASL auth over TLS (recommended for production)
type Config struct {
	BootstrapServers string
	SecurityProtocol string // PLAINTEXT | SSL | SASL_PLAINTEXT | SASL_SSL
	SASLMechanism    string // SCRAM-SHA-256 | SCRAM-SHA-512
	SASLUsername     string
	SASLPassword     string
	TLSCACertPath    string // Path to CA certificate (PEM) for broker verification
	TLSCertPath      string // Path to client certificate (PEM) for mTLS
	TLSKeyPath       string // Path to client private key (PEM) for mTLS
	TLSSkipVerify    bool   // Skip server certificate verification (dev/test only)
	Region           string
}

// Client provides Kafka producer and consumer capabilities
type Client struct {
	config  *Config
	dialer  *kafka.Dialer
	writers map[string]*kafka.Writer
}

// NewConfigFromEnv creates Config from environment variables.
func NewConfigFromEnv() *Config {
	return &Config{
		BootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		SecurityProtocol: getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
		SASLMechanism:    getEnv("KAFKA_SASL_MECHANISM", ""),
		SASLUsername:     getEnv("KAFKA_SASL_USERNAME", ""),
		SASLPassword:     getEnv("KAFKA_SASL_PASSWORD", ""),
		TLSCACertPath:    getEnv("KAFKA_TLS_CA_CERT_PATH", ""),
		TLSCertPath:      getEnv("KAFKA_TLS_CERT_PATH", ""),
		TLSKeyPath:       getEnv("KAFKA_TLS_KEY_PATH", ""),
		TLSSkipVerify:    getEnv("KAFKA_TLS_SKIP_VERIFY", "false") == "true",
		Region:           getEnv("REGION", "local"),
	}
}

// NewClient creates a new Kafka client with the given configuration.
// Supports PLAINTEXT, SSL, SASL_PLAINTEXT, and SASL_SSL security protocols.
func NewClient(config *Config) (*Client, error) {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// Configure SASL authentication if mechanism is set
	if config.SASLMechanism != "" {
		mechanism, err := buildSASLMechanism(config)
		if err != nil {
			return nil, err
		}
		dialer.SASLMechanism = mechanism
	}

	// Configure TLS for SSL or SASL_SSL protocols
	if config.SecurityProtocol == "SSL" || config.SecurityProtocol == "SASL_SSL" {
		tlsConfig, err := buildTLSConfig(config)
		if err != nil {
			return nil, err
		}
		dialer.TLS = tlsConfig
	}

	slog.Info("Kafka client configured",
		"bootstrap_servers", config.BootstrapServers,
		"security_protocol", config.SecurityProtocol,
		"sasl_mechanism", config.SASLMechanism,
		"tls_ca_cert", config.TLSCACertPath != "",
		"tls_client_cert", config.TLSCertPath != "",
		"region", config.Region)

	return &Client{
		config:  config,
		dialer:  dialer,
		writers: make(map[string]*kafka.Writer),
	}, nil
}

// buildSASLMechanism creates the SASL mechanism from config.
func buildSASLMechanism(config *Config) (sasl.Mechanism, error) {
	switch config.SASLMechanism {
	case "SCRAM-SHA-512":
		mechanism, err := scram.Mechanism(scram.SHA512, config.SASLUsername, config.SASLPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to create SCRAM-SHA-512 mechanism: %w", err)
		}
		return mechanism, nil
	case "SCRAM-SHA-256":
		mechanism, err := scram.Mechanism(scram.SHA256, config.SASLUsername, config.SASLPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to create SCRAM-SHA-256 mechanism: %w", err)
		}
		return mechanism, nil
	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s (supported: SCRAM-SHA-256, SCRAM-SHA-512)", config.SASLMechanism)
	}
}

// buildTLSConfig creates TLS configuration from config.
// Supports custom CA certificate, client certificates (mTLS), and skip-verify for dev.
func buildTLSConfig(config *Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: config.TLSSkipVerify,
	}

	// Load custom CA certificate for broker verification
	if config.TLSCACertPath != "" {
		caCert, err := os.ReadFile(config.TLSCACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate from %s: %w", config.TLSCACertPath, err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate from %s", config.TLSCACertPath)
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key for mTLS
	if config.TLSCertPath != "" && config.TLSKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(config.TLSCertPath, config.TLSKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate/key (%s, %s): %w",
				config.TLSCertPath, config.TLSKeyPath, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// GetWriter returns a cached writer for the given topic
func (c *Client) GetWriter(topic string) *kafka.Writer {
	if writer, exists := c.writers[topic]; exists {
		return writer
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(c.config.BootstrapServers, ",")...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		Transport: &kafka.Transport{
			Dial: c.dialer.DialFunc,
			SASL: c.dialer.SASLMechanism,
			TLS:  c.dialer.TLS,
		},
	}

	c.writers[topic] = writer
	return writer
}

// Close closes all writers
func (c *Client) Close() error {
	var errs []error
	for _, writer := range c.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing writers: %v", errs)
	}
	return nil
}

// Brokers returns the list of broker addresses
func (c *Client) Brokers() []string {
	return strings.Split(c.config.BootstrapServers, ",")
}

// Dialer returns the configured dialer for creating readers
func (c *Client) Dialer() *kafka.Dialer {
	return c.dialer
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Message represents a Kafka message with metadata
type Message struct {
	Key       string
	Value     interface{}
	Headers   map[string]string
	Timestamp time.Time
}

// Publish sends a message to the specified topic
func (c *Client) Publish(ctx context.Context, topic string, msg *Message) error {
	value, err := json.Marshal(msg.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	headers := make([]kafka.Header, 0, len(msg.Headers)+1)
	for k, v := range msg.Headers {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	headers = append(headers, kafka.Header{Key: "region", Value: []byte(c.config.Region)})

	kafkaMsg := kafka.Message{
		Key:     []byte(msg.Key),
		Value:   value,
		Headers: headers,
		Time:    msg.Timestamp,
	}

	writer := c.GetWriter(topic)
	if err := writer.WriteMessages(ctx, kafkaMsg); err != nil {
		slog.Error("Failed to publish message",
			"topic", topic,
			"key", msg.Key,
			"error", err)
		return fmt.Errorf("failed to write message: %w", err)
	}

	slog.Debug("Published message",
		"topic", topic,
		"key", msg.Key,
		"region", c.config.Region)

	return nil
}

// PublishBatch sends multiple messages to the specified topic
func (c *Client) PublishBatch(ctx context.Context, topic string, msgs []*Message) error {
	kafkaMsgs := make([]kafka.Message, len(msgs))
	for i, msg := range msgs {
		value, err := json.Marshal(msg.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal message %d: %w", i, err)
		}

		headers := make([]kafka.Header, 0, len(msg.Headers)+1)
		for k, v := range msg.Headers {
			headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
		}
		headers = append(headers, kafka.Header{Key: "region", Value: []byte(c.config.Region)})

		kafkaMsgs[i] = kafka.Message{
			Key:     []byte(msg.Key),
			Value:   value,
			Headers: headers,
			Time:    msg.Timestamp,
		}
	}

	writer := c.GetWriter(topic)
	if err := writer.WriteMessages(ctx, kafkaMsgs...); err != nil {
		slog.Error("Failed to publish batch",
			"topic", topic,
			"count", len(msgs),
			"error", err)
		return fmt.Errorf("failed to write batch: %w", err)
	}

	slog.Debug("Published batch",
		"topic", topic,
		"count", len(msgs),
		"region", c.config.Region)

	return nil
}
