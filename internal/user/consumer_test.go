package user

import (
	"context"
	"errors"
	"testing"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
)

type mockUserRepo struct {
	deactivatedUsers []string
	deletedUsers     []string
	errToReturn      error
}

func (m *mockUserRepo) DeactivateUser(ctx context.Context, id string) error {
	if m.errToReturn != nil {
		return m.errToReturn
	}
	m.deactivatedUsers = append(m.deactivatedUsers, id)
	return nil
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, id string) error {
	if m.errToReturn != nil {
		return m.errToReturn
	}
	m.deletedUsers = append(m.deletedUsers, id)
	return nil
}

type mockRefreshRepo struct {
	revokedUsers []string
	errToReturn  error
}

func (m *mockRefreshRepo) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	if m.errToReturn != nil {
		return m.errToReturn
	}
	m.revokedUsers = append(m.revokedUsers, userID)
	return nil
}

type mockKafkaConsumer struct {
	messagesToFetch []kafkaGo.Message
	fetchIndex      int
	committed       []kafkaGo.Message
	closed          bool
}

func (m *mockKafkaConsumer) FetchMessage(ctx context.Context) (kafkaGo.Message, error) {
	if m.fetchIndex < len(m.messagesToFetch) {
		msg := m.messagesToFetch[m.fetchIndex]
		m.fetchIndex++
		return msg, nil
	}
	<-ctx.Done()
	return kafkaGo.Message{}, ctx.Err()
}

func (m *mockKafkaConsumer) CommitMessages(ctx context.Context, msgs ...kafkaGo.Message) error {
	m.committed = append(m.committed, msgs...)
	return nil
}

func (m *mockKafkaConsumer) Close() error {
	m.closed = true
	return nil
}

func TestUserBannedEvent(t *testing.T) {
	userRepo := &mockUserRepo{}
	refreshRepo := &mockRefreshRepo{}
	mockKafka := &mockKafkaConsumer{
		messagesToFetch: []kafkaGo.Message{
			{
				Topic: "user.events",
				Key:   []byte("user-123"),
				Value: []byte(`{"event":"user.banned","userId":"user-123","reason":"Terms violation"}`),
			},
		},
	}

	consumer := NewConsumer(mockKafka, userRepo, refreshRepo, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	consumer.Start(ctx)

	if len(userRepo.deactivatedUsers) != 1 || userRepo.deactivatedUsers[0] != "user-123" {
		t.Errorf("expected user-123 to be deactivated; got %v", userRepo.deactivatedUsers)
	}

	if len(refreshRepo.revokedUsers) != 1 || refreshRepo.revokedUsers[0] != "user-123" {
		t.Errorf("expected refresh tokens for user-123 to be revoked; got %v", refreshRepo.revokedUsers)
	}

	if len(mockKafka.committed) != 1 {
		t.Errorf("expected 1 message committed; got %d", len(mockKafka.committed))
	}
}

func TestUserDeletedEvent(t *testing.T) {
	userRepo := &mockUserRepo{}
	refreshRepo := &mockRefreshRepo{}
	mockKafka := &mockKafkaConsumer{
		messagesToFetch: []kafkaGo.Message{
			{
				Topic: "user.events",
				Key:   []byte("user-456"),
				Value: []byte(`{"event":"user.deleted","userId":"user-456"}`),
			},
		},
	}

	consumer := NewConsumer(mockKafka, userRepo, refreshRepo, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	consumer.Start(ctx)

	if len(userRepo.deletedUsers) != 1 || userRepo.deletedUsers[0] != "user-456" {
		t.Errorf("expected user-456 to be deleted; got %v", userRepo.deletedUsers)
	}

	if len(refreshRepo.revokedUsers) != 1 || refreshRepo.revokedUsers[0] != "user-456" {
		t.Errorf("expected refresh tokens for user-456 to be revoked; got %v", refreshRepo.revokedUsers)
	}

	if len(mockKafka.committed) != 1 {
		t.Errorf("expected 1 message committed; got %d", len(mockKafka.committed))
	}
}

func TestMalformedAndUnknownEvents(t *testing.T) {
	userRepo := &mockUserRepo{}
	refreshRepo := &mockRefreshRepo{}
	mockKafka := &mockKafkaConsumer{
		messagesToFetch: []kafkaGo.Message{
			{
				Topic: "user.events",
				Value: []byte(`not-json`),
			},
			{
				Topic: "user.events",
				Value: []byte(`{"event":"user.banned"}`), // missing userId
			},
			{
				Topic: "user.events",
				Value: []byte(`{"event":"user.unknown","userId":"user-789"}`),
			},
		},
	}

	consumer := NewConsumer(mockKafka, userRepo, refreshRepo, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	consumer.Start(ctx)

	if len(userRepo.deactivatedUsers) != 0 {
		t.Errorf("expected no users deactivated; got %v", userRepo.deactivatedUsers)
	}
	if len(userRepo.deletedUsers) != 0 {
		t.Errorf("expected no users deleted; got %v", userRepo.deletedUsers)
	}
	if len(mockKafka.committed) != 3 {
		t.Errorf("expected all 3 messages committed to prevent blocking; got %d", len(mockKafka.committed))
	}
}

func TestRepositoryErrorsAreNonFatal(t *testing.T) {
	userRepo := &mockUserRepo{errToReturn: errors.New("db connection failure")}
	refreshRepo := &mockRefreshRepo{errToReturn: errors.New("db error")}
	mockKafka := &mockKafkaConsumer{
		messagesToFetch: []kafkaGo.Message{
			{
				Topic: "user.events",
				Value: []byte(`{"event":"user.banned","userId":"user-err"}`),
			},
		},
	}

	consumer := NewConsumer(mockKafka, userRepo, refreshRepo, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	consumer.Start(ctx)

	if len(mockKafka.committed) != 1 {
		t.Errorf("expected message committed even with repository error; got %d", len(mockKafka.committed))
	}
}

func TestConsumerClose(t *testing.T) {
	mockKafka := &mockKafkaConsumer{}
	consumer := NewConsumer(mockKafka, nil, nil, nil)

	if err := consumer.Close(); err != nil {
		t.Fatalf("unexpected error closing consumer: %v", err)
	}

	if !mockKafka.closed {
		t.Error("expected underlying kafka consumer to be closed")
	}
}
