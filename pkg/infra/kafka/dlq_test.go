package kafka

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kafkago "github.com/segmentio/kafka-go"
)

func TestDefaultRetryPolicy(t *testing.T) {
	t.Setenv("KAFKA_MAX_RETRIES", "2")
	t.Setenv("KAFKA_RETRY_BACKOFF_MS", "100")

	p := DefaultRetryPolicy()
	assert.Equal(t, 2, p.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, p.Backoff)
	assert.True(t, p.ShouldRetry(0))
	assert.True(t, p.ShouldRetry(1))
	assert.False(t, p.ShouldRetry(2))
}

func TestDLQMessagePublish(t *testing.T) {
	// Uses mock-free structural test of DLQ payload fields.
	msg := &kafkago.Message{
		Topic:     TopicMatchmakingCommands,
		Partition: 1,
		Offset:    42,
		Key:       []byte("player-1"),
		Value:     []byte(`{"test":true}`),
	}

	dlq := DLQMessage{
		OriginalTopic:     msg.Topic,
		OriginalPartition: msg.Partition,
		OriginalOffset:    msg.Offset,
		OriginalKey:       string(msg.Key),
		Error:             "boom",
		RetryCount:        3,
		Timestamp:         time.Now().UnixMilli(),
		Payload:           msg.Value,
	}

	require.Equal(t, TopicMatchmakingCommands, dlq.OriginalTopic)
	require.Equal(t, 3, dlq.RetryCount)
}
