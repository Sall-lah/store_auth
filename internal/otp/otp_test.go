package otp

import (
	"strconv"
	"testing"
)

func TestOTP(t *testing.T) {
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

	t.Run("NewOTPSender creates correct sender implementation", func(t *testing.T) {
		mockSender := NewOTPSender("mock", "", 0, "", "", "")
		if _, ok := mockSender.(*LogOTPSender); !ok {
			t.Errorf("expected *LogOTPSender; got %T", mockSender)
		}

		if err := mockSender.Send("user@example.com", "123456"); err != nil {
			t.Errorf("expected mock send to succeed; got %v", err)
		}

		smtpSender := NewOTPSender("smtp", "smtp.gmail.com", 587, "user", "pass", "noreply@store.com")
		if _, ok := smtpSender.(*SMTPOTPSender); !ok {
			t.Errorf("expected *SMTPOTPSender; got %T", smtpSender)
		}
	})
}
