package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
)

// Consumer defines the interface for consuming messages from a Kafka consumer group.
// Why: Decouples message processing logic from specific Kafka transport implementations and enables unit test mocking.
type Consumer interface {
	FetchMessage(ctx context.Context) (kafkaGo.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafkaGo.Message) error
	Close() error
}

// KafkaConsumer wraps segmentio/kafka-go Reader for pure Go consumer group message processing.
type KafkaConsumer struct {
	reader *kafkaGo.Reader
}

// NewConsumer constructs a pure Go KafkaConsumer configured with consumer group rebalancing and smart dialer resilience.
// Why: Provides non-blocking, partition-balanced Kafka message consumption without external C runtime dependencies.
func NewConsumer(brokers []string, topic, groupID string) *KafkaConsumer {
	dialer := newSmartDialer()

	reader := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		Dialer:         dialer,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: 0,    // Synchronous/explicit commit per message batch
		StartOffset:    kafkaGo.FirstOffset,
		MaxWait:        500 * time.Millisecond,
		ErrorLogger: kafkaGo.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("[Kafka Consumer Error] "+msg, args...)
		}),
	})

	return &KafkaConsumer{reader: reader}
}

// FetchMessage retrieves the next available message from the assigned Kafka partition.
// Why: Enables explicit per-message processing loops with full context cancellation support.
func (c *KafkaConsumer) FetchMessage(ctx context.Context) (kafkaGo.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafkaGo.Message{}, fmt.Errorf("failed to fetch message from kafka: %w", err)
	}
	return msg, nil
}

// CommitMessages marks message offsets as processed in the consumer group coordinator.
// Why: Guarantees at-least-once message processing and avoids reprocessing already handled lifecycle events.
func (c *KafkaConsumer) CommitMessages(ctx context.Context, msgs ...kafkaGo.Message) error {
	if err := c.reader.CommitMessages(ctx, msgs...); err != nil {
		return fmt.Errorf("failed to commit kafka message offsets: %w", err)
	}
	return nil
}

// Close disconnects from the Kafka cluster and releases consumer group partition assignments.
// Why: Ensures graceful partition rebalancing during service restarts or scale-downs.
func (c *KafkaConsumer) Close() error {
	log.Println("[Kafka] Closing Kafka consumer reader...")
	return c.reader.Close()
}
