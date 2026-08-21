package otp

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
)

var (
	ErrOTPExpired     = errors.New("otp code has expired")
	ErrOTPInvalidCode = errors.New("invalid otp code")
	ErrOTPAlreadyUsed = errors.New("otp code already used")
	ErrOTPMaxAttempts = errors.New("maximum otp verification attempts exceeded")
)

// Service orchestrates generation, persistence, delivery, and verification of 2FA verification codes.
type Service struct {
	otpRepo       *Repository
	sender        OTPSender
	expiryMinutes int
	maxAttempts   int
}

// NewService constructs an OTP Service initialized with expiry rules and delivery provider.
func NewService(otpRepo *Repository, sender OTPSender, expiryMinutes, maxAttempts int) *Service {
	if sender == nil {
		sender = &LogOTPSender{}
	}
	return &Service{
		otpRepo:       otpRepo,
		sender:        sender,
		expiryMinutes: expiryMinutes,
		maxAttempts:   maxAttempts,
	}
}

// GenerateOTP creates a cryptographically secure 6-digit numeric verification code and persists it.
func (s *Service) GenerateOTP(ctx context.Context, userID string, otpType Type) (*OTPCode, error) {
	code, err := generateRandomNumericCode(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random otp code: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(s.expiryMinutes) * time.Minute)

	otp, err := s.otpRepo.CreateOTP(ctx, userID, code, otpType, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to persist otp: %w", err)
	}

	return otp, nil
}

// SendOTP triggers delivery of the generated OTP code to the target destination via the configured OTPSender interface.
// Why: Delegates notification event publishing and formatting to the configured transport provider.
func (s *Service) SendOTP(ctx context.Context, email, code, name string, otpType Type) error {
	return s.sender.SendOTP(ctx, email, code, name, otpType)
}

// VerifyOTP validates an incoming user-entered OTP against the stored active record for that flow.
func (s *Service) VerifyOTP(ctx context.Context, userID, code string, otpType Type) error {
	otpRecord, err := s.otpRepo.FindLatestOTPByUserAndType(ctx, userID, otpType)
	if err != nil {
		if errors.Is(err, ErrOTPNotFound) {
			return ErrOTPInvalidCode
		}
		return err
	}

	if otpRecord.Used {
		return ErrOTPAlreadyUsed
	}

	if otpRecord.Attempts >= s.maxAttempts {
		return ErrOTPMaxAttempts
	}

	if time.Now().After(otpRecord.ExpiresAt) {
		return ErrOTPExpired
	}

	if otpRecord.Code != code {
		_ = s.otpRepo.IncrementOTPAttempts(ctx, otpRecord.ID)
		return ErrOTPInvalidCode
	}

	if err := s.otpRepo.MarkOTPUsed(ctx, otpRecord.ID); err != nil {
		return fmt.Errorf("failed to mark otp as used: %w", err)
	}

	return nil
}

// InvalidateUserOTPs revokes remaining pending codes for a user upon successful reset or account updates.
func (s *Service) InvalidateUserOTPs(ctx context.Context, userID string, otpType Type) error {
	return s.otpRepo.InvalidateOTPsByUserAndType(ctx, userID, otpType)
}

func generateRandomNumericCode(length int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[num.Int64()]
	}
	return string(result), nil
}
