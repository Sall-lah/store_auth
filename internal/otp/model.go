package otp

import (
	"time"
)

// Type distinguishes between user registration activation and password recovery verification flows.
type Type string

const (
	TypeRegistration  Type = "REGISTRATION"
	TypePasswordReset Type = "PASSWORD_RESET"
)

// OTPCode represents a temporary verification code issued during 2FA/activation.
type OTPCode struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	Type      Type      `json:"type"`
	Attempts  int       `json:"attempts"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// VerifyOTPRequest contains verification details to confirm account registration or password reset authorization.
type VerifyOTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
	Type  Type   `json:"type"`
}

// ResetPasswordRequest contains the authorization code and new password string.
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}
