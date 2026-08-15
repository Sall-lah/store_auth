package otp

import (
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"strings"
)

// OTPSender abstracts out external SMS/Email OTP transport mechanisms, allowing environment-specific senders.
type OTPSender interface {
	Send(destination, code string) error
}

// LogOTPSender is a mock/development implementation of OTPSender that logs OTP codes to standard output.
type LogOTPSender struct{}

// Send logs the generated OTP code to console for local testing without external gateway costs.
func (s *LogOTPSender) Send(destination, code string) error {
	log.Printf("[OTP DEV LOG] Sending OTP Code '%s' to target: '%s'", code, destination)
	return nil
}

// SMTPOTPSender delivers OTP verification codes via SMTP email transport.
type SMTPOTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewSMTPOTPSender constructs an SMTP sender instance configured with target relay server credentials.
func NewSMTPOTPSender(host string, port int, username, password, from string) *SMTPOTPSender {
	if from == "" {
		from = username
	}
	return &SMTPOTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// Send transmits an HTML formatted OTP verification email to the destination address via SMTP.
func (s *SMTPOTPSender) Send(destination, code string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	envelopeFrom := s.username
	if parsed, err := mail.ParseAddress(s.from); err == nil {
		envelopeFrom = parsed.Address
	} else if s.from != "" {
		envelopeFrom = s.from
	}

	subject := "Your Verification Code"
	fromHeader := s.from
	if fromHeader == "" {
		fromHeader = s.username
	}

	body := fmt.Sprintf("Subject: %s\r\n"+
		"From: %s\r\n"+
		"To: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		"<div style=\"font-family: Arial, sans-serif; padding: 20px;\">"+
		"<h2>Your Verification Code</h2>"+
		"<p>Use the following 6-digit code to complete your verification:</p>"+
		"<h1 style=\"font-size: 32px; letter-spacing: 5px; color: #2563eb;\">%s</h1>"+
		"<p>This code expires in 5 minutes. If you did not request this code, please ignore this email.</p>"+
		"</div>", subject, fromHeader, destination, code)

	err := smtp.SendMail(addr, auth, envelopeFrom, []string{destination}, []byte(body))
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	log.Printf("[OTP SMTP LOG] Successfully sent OTP code to '%s'", destination)
	return nil
}

// NewOTPSender constructs an appropriate OTPSender instance based on the specified provider strategy ("smtp" vs "mock").
func NewOTPSender(provider, host string, port int, username, password, from string) OTPSender {
	if strings.ToLower(provider) == "smtp" {
		return NewSMTPOTPSender(host, port, username, password, from)
	}
	return &LogOTPSender{}
}
