package otp

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"store_auth/internal/platform/kafka"
)

// OTPSender abstracts notification transport mechanisms for OTP delivery.
// Why: Enables hot-swappable delivery mechanisms (e.g. Kafka event dispatching in production, stdout logging in tests).
type OTPSender interface {
	SendOTP(ctx context.Context, email, code, name string, otpType Type) error
}

// LogOTPSender is a mock/development implementation of OTPSender that logs OTP codes to standard output.
type LogOTPSender struct{}

// SendOTP logs the generated OTP code and metadata to console for local testing without external broker dependencies.
// Why: Provides lightweight offline debugging and test verification capabilities.
func (s *LogOTPSender) SendOTP(ctx context.Context, email, code, name string, otpType Type) error {
	log.Printf("[OTP DEV LOG] Sending %s OTP Code '%s' to '%s' (User: '%s')", otpType, code, email, name)
	return nil
}

// EventEnvelope models the standard message envelope format consumed by store_notification.
type EventEnvelope struct {
	EventID   string           `json:"event_id"`
	EventType string           `json:"event_type"`
	Timestamp time.Time        `json:"timestamp"`
	Producer  string           `json:"producer"`
	Data      AuthOtpEventData `json:"data"`
}

// AuthOtpEventData models the payload consumed by store_notification for rendering OTP email templates.
type AuthOtpEventData struct {
	Email string `json:"email"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

// KafkaOTPSender delivers OTP verification events to store_notification via Apache Kafka.
type KafkaOTPSender struct {
	producer kafka.Producer
	topic    string
}

// NewKafkaOTPSender constructs a KafkaOTPSender bound to a Kafka producer client and destination topic.
// Why: Injects Kafka message broker dependencies to publish auth lifecycle domain events.
func NewKafkaOTPSender(producer kafka.Producer, topic string) *KafkaOTPSender {
	if topic == "" {
		topic = "auth.events"
	}
	return &KafkaOTPSender{
		producer: producer,
		topic:    topic,
	}
}

// SendOTP creates a validated EventEnvelope and publishes it to the configured Kafka topic.
// Why: Dispatches asynchronous events to store_notification for template rendering and email transmission.
func (s *KafkaOTPSender) SendOTP(ctx context.Context, email, code, name string, otpType Type) error {
	eventID, err := generateUUID()
	if err != nil {
		return fmt.Errorf("failed to generate event id: %w", err)
	}

	eventType := "auth.registration_otp"
	if otpType == TypePasswordReset {
		eventType = "auth.password_reset_otp"
	}

	envelope := EventEnvelope{
		EventID:   eventID,
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Producer:  "store_auth",
		Data: AuthOtpEventData{
			Email: email,
			Code:  code,
			Name:  name,
			Type:  string(otpType),
		},
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal otp event envelope: %w", err)
	}

	if err := s.producer.Publish(ctx, s.topic, email, payload); err != nil {
		return fmt.Errorf("failed to publish otp event to kafka topic %s: %w", s.topic, err)
	}

	log.Printf("[OTP KAFKA LOG] Successfully published '%s' event for '%s' to topic '%s'", eventType, email, s.topic)
	return nil
}

// NewOTPSender constructs an appropriate OTPSender instance based on the specified provider strategy ("kafka" vs "mock").
// Why: Simplifies dependency injection by selecting the configured OTP delivery implementation at runtime.
func NewOTPSender(provider string, producer kafka.Producer, topic string) OTPSender {
	if strings.ToLower(provider) == "kafka" && producer != nil {
		return NewKafkaOTPSender(producer, topic)
	}
	return &LogOTPSender{}
}

// generateUUID creates an RFC 4122 v4 compliant UUID using cryptographically secure random bytes.
// Why: Guarantees unique event IDs for Redis idempotency deduplication in store_notification without external CGo dependencies.
func generateUUID() (string, error) {
	var uuid [16]byte
	if _, err := rand.Read(uuid[:]); err != nil {
		return "", err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
