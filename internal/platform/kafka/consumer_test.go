package kafka

import (
	"context"
	"testing"

	kafkaGo "github.com/segmentio/kafka-go"
)

type mockConsumer struct {
	messagesToFetch []kafkaGo.Message
	fetchIndex      int
	committed       []kafkaGo.Message
	closed          bool
}

func (m *mockConsumer) FetchMessage(ctx context.Context) (kafkaGo.Message, error) {
	if m.fetchIndex < len(m.messagesToFetch) {
		msg := m.messagesToFetch[m.fetchIndex]
		m.fetchIndex++
		return msg, nil
	}
	<-ctx.Done()
	return kafkaGo.Message{}, ctx.Err()
}

func (m *mockConsumer) CommitMessages(ctx context.Context, msgs ...kafkaGo.Message) error {
	m.committed = append(m.committed, msgs...)
	return nil
}

func (m *mockConsumer) Close() error {
	m.closed = true
	return nil
}

func TestMockConsumerInterface(t *testing.T) {
	mockMsg := kafkaGo.Message{
		Topic: "user.events",
		Key:   []byte("test-user-id"),
		Value: []byte(`{"event":"user.banned","userId":"test-user-id"}`),
	}

	var c Consumer = &mockConsumer{
		messagesToFetch: []kafkaGo.Message{mockMsg},
	}

	ctx := context.Background()
	msg, err := c.FetchMessage(ctx)
	if err != nil {
		t.Fatalf("unexpected error fetching message: %v", err)
	}

	if string(msg.Value) != string(mockMsg.Value) {
		t.Errorf("expected payload %s; got %s", string(mockMsg.Value), string(msg.Value))
	}

	if err := c.CommitMessages(ctx, msg); err != nil {
		t.Fatalf("unexpected error committing messages: %v", err)
	}

	mc := c.(*mockConsumer)
	if len(mc.committed) != 1 {
		t.Errorf("expected 1 committed message; got %d", len(mc.committed))
	}

	if err := c.Close(); err != nil {
		t.Fatalf("unexpected error closing consumer: %v", err)
	}
	if !mc.closed {
		t.Error("expected mock consumer to be marked closed")
	}
}

func TestNewConsumerConstructor(t *testing.T) {
	brokers := []string{"localhost:9092", "broker:9092"}
	consumer := NewConsumer(brokers, "user.events", "store-auth-user-events-group")

	if consumer == nil {
		t.Fatal("expected NewConsumer to return non-nil instance")
	}
	if consumer.reader == nil {
		t.Fatal("expected reader to be initialized")
	}
}
