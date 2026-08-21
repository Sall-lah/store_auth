package kafka

import (
	"context"
	"testing"
)

type mockProducer struct {
	publishedTopic   string
	publishedKey     string
	publishedPayload []byte
	closed           bool
}

func (m *mockProducer) Publish(ctx context.Context, topic, key string, payload []byte) error {
	m.publishedTopic = topic
	m.publishedKey = key
	m.publishedPayload = payload
	return nil
}

func (m *mockProducer) Close() error {
	m.closed = true
	return nil
}

func TestMockProducerInterface(t *testing.T) {
	var p Producer = &mockProducer{}

	ctx := context.Background()
	topic := "auth.events"
	key := "user@example.com"
	payload := []byte(`{"event_id":"test-123"}`)

	if err := p.Publish(ctx, topic, key, payload); err != nil {
		t.Fatalf("unexpected error publishing: %v", err)
	}

	mp := p.(*mockProducer)
	if mp.publishedTopic != topic {
		t.Errorf("expected topic %s; got %s", topic, mp.publishedTopic)
	}
	if mp.publishedKey != key {
		t.Errorf("expected key %s; got %s", key, mp.publishedKey)
	}
	if string(mp.publishedPayload) != string(payload) {
		t.Errorf("expected payload %s; got %s", string(payload), string(mp.publishedPayload))
	}

	if err := p.Close(); err != nil {
		t.Fatalf("unexpected error closing producer: %v", err)
	}
	if !mp.closed {
		t.Error("expected mock producer to be marked closed")
	}
}

func TestNewProducerConstructor(t *testing.T) {
	brokers := []string{"localhost:9092", "broker:9092"}
	producer := NewProducer(brokers)

	if producer == nil {
		t.Fatal("expected NewProducer to return non-nil instance")
	}
	if producer.writer == nil {
		t.Fatal("expected writer to be initialized")
	}
}
