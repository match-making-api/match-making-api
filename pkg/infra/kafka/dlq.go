package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// DLQMessage is the canonical dead-letter payload (Epic §10).
type DLQMessage struct {
	OriginalTopic     string `json:"original_topic"`
	OriginalPartition int    `json:"original_partition"`
	OriginalOffset    int64  `json:"original_offset"`
	OriginalKey       string `json:"original_key"`
	Error             string `json:"error"`
	RetryCount        int    `json:"retry_count"`
	Timestamp         int64  `json:"timestamp"`
	Payload           []byte `json:"payload"`
}

// DLQPublisher sends failed messages to matchmaking.dlq.
type DLQPublisher struct {
	client *Client
}

// NewDLQPublisher creates a DLQ publisher.
func NewDLQPublisher(client *Client) *DLQPublisher {
	return &DLQPublisher{client: client}
}

// Publish writes a DLQ record. Offset is not committed for the original message when DLQ succeeds.
func (p *DLQPublisher) Publish(ctx context.Context, msg *kafka.Message, retryCount int, processErr error) error {
	dlq := DLQMessage{
		OriginalTopic:     msg.Topic,
		OriginalPartition: msg.Partition,
		OriginalOffset:    msg.Offset,
		OriginalKey:       string(msg.Key),
		Error:             processErr.Error(),
		RetryCount:        retryCount,
		Timestamp:         time.Now().UnixMilli(),
		Payload:           append([]byte(nil), msg.Value...),
	}

	body, err := json.Marshal(dlq)
	if err != nil {
		return fmt.Errorf("marshal dlq message: %w", err)
	}

	key := dlq.OriginalKey
	if key == "" {
		key = fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Partition, msg.Offset)
	}

	headers := map[string]string{
		"original_topic": msg.Topic,
		"error_type":     "processing_failed",
		"retry_count":    fmt.Sprintf("%d", retryCount),
	}

	return p.client.PublishBytes(ctx, TopicDLQ, key, body, headers)
}
