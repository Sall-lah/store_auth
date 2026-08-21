package otp

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
)

type testMockProducer struct {
	publishedTopic   string
	publishedKey     string
	publishedPayload []byte
}

func (m *testMockProducer) Publish(ctx context.Context, topic, key string, payload []byte) error {
	m.publishedTopic = topic
	m.publishedKey = key
	m.publishedPayload = payload
	return nil
}

func (m *testMockProducer) Close() error {
	return nil
}

func TestOTP(t *testing.T) {
	ctx := context.Background()

	t.Run("generateRandomNumericCode generates 6-digit string", func(t *testing.T) {
		code, err := generateRandomNumericCode(6)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 6 {
			t.Errorf("expected 6-digit code; got %d characters: %s", len(code), code)
		}
		if _, err := strconv.Atoi(code); err != nil {
			t.Errorf("expected numeric string; got %s", code)
		}
	})

	t.Run("generateUUID generates valid v4 UUID format", func(t *testing.T) {
		id, err := generateUUID()
		if err != nil {
			t.Fatalf("unexpected error generating uuid: %v", err)
		}
		if len(id) != 36 {
			t.Errorf("expected 36 characters; got %d (%s)", len(id), id)
		}
		if id[14] != '4' {
			t.Errorf("expected UUID version 4; got %c in %s", id[14], id)
		}
	})

	t.Run("LogOTPSender logs without error", func(t *testing.T) {
		sender := &LogOTPSender{}
		err := sender.SendOTP(ctx, "user@example.com", "123456", "John Doe", TypeRegistration)
		if err != nil {
			t.Errorf("expected nil error; got %v", err)
		}
	})

	t.Run("KafkaOTPSender publishes valid EventEnvelope for Registration", func(t *testing.T) {
		producer := &testMockProducer{}
		sender := NewKafkaOTPSender(producer, "auth.events")

		err := sender.SendOTP(ctx, "customer@example.com", "654321", "Jane Doe", TypeRegistration)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if producer.publishedTopic != "auth.events" {
			t.Errorf("expected topic auth.events; got %s", producer.publishedTopic)
		}
		if producer.publishedKey != "customer@example.com" {
			t.Errorf("expected key customer@example.com; got %s", producer.publishedKey)
		}

		var envelope EventEnvelope
		if err := json.Unmarshal(producer.publishedPayload, &envelope); err != nil {
			t.Fatalf("failed to unmarshal payload envelope: %v", err)
		}

		if envelope.EventType != "auth.registration_otp" {
			t.Errorf("expected event_type 'auth.registration_otp'; got %s", envelope.EventType)
		}
		if envelope.Producer != "store_auth" {
			t.Errorf("expected producer 'store_auth'; got %s", envelope.Producer)
		}
		if envelope.Data.Code != "654321" {
			t.Errorf("expected code '654321'; got %s", envelope.Data.Code)
		}
		if envelope.Data.Name != "Jane Doe" {
			t.Errorf("expected name 'Jane Doe'; got %s", envelope.Data.Name)
		}
		if envelope.Data.Type != string(TypeRegistration) {
			t.Errorf("expected data type 'REGISTRATION'; got %s", envelope.Data.Type)
		}
	})

	t.Run("KafkaOTPSender publishes valid EventEnvelope for Password Reset", func(t *testing.T) {
		producer := &testMockProducer{}
		sender := NewKafkaOTPSender(producer, "auth.events")

		err := sender.SendOTP(ctx, "customer@example.com", "999888", "Jane Doe", TypePasswordReset)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var envelope EventEnvelope
		if err := json.Unmarshal(producer.publishedPayload, &envelope); err != nil {
			t.Fatalf("failed to unmarshal payload envelope: %v", err)
		}

		if envelope.EventType != "auth.password_reset_otp" {
			t.Errorf("expected event_type 'auth.password_reset_otp'; got %s", envelope.EventType)
		}
		if envelope.Data.Type != string(TypePasswordReset) {
			t.Errorf("expected data type 'PASSWORD_RESET'; got %s", envelope.Data.Type)
		}
	})

	t.Run("NewOTPSender creates correct sender implementation", func(t *testing.T) {
		mockSender := NewOTPSender("mock", nil, "")
		if _, ok := mockSender.(*LogOTPSender); !ok {
			t.Errorf("expected *LogOTPSender; got %T", mockSender)
		}

		producer := &testMockProducer{}
		kafkaSender := NewOTPSender("kafka", producer, "auth.events")
		if _, ok := kafkaSender.(*KafkaOTPSender); !ok {
			t.Errorf("expected *KafkaOTPSender; got %T", kafkaSender)
		}
	})
}
