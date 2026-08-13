package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"store_auth/internal/jwt"
	"store_auth/internal/otp"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountInactive    = errors.New("user account is inactive, please verify registration OTP")
	ErrEmailTaken         = errors.New("email address is already registered")
)

// Service encapsulates authentication logic, credential processing, and feature integration.
type Service struct {
	userRepo   *Repository
	otpService *otp.Service
	jwtService *jwt.Service
	bcryptCost int
}

// NewService constructs a Service binding user repository, JWT signer, and OTP manager.
func NewService(userRepo *Repository, otpService *otp.Service, jwtService *jwt.Service, bcryptCost int) *Service {
	return &Service{
		userRepo:   userRepo,
		otpService: otpService,
		jwtService: jwtService,
		bcryptCost: bcryptCost,
	}
}

// Register creates an inactive user account, generates a registration verification OTP, and triggers transmission.
func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	existing, err := s.userRepo.FindUserByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return ErrEmailTaken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.userRepo.CreateUser(ctx, req.Email, string(hashedPassword), req.Name, RoleCustomer)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return ErrEmailTaken
		}
		return err
	}

	otpCode, err := s.otpService.GenerateOTP(ctx, user.ID, otp.TypeRegistration)
	if err != nil {
		return fmt.Errorf("failed to generate registration otp: %w", err)
	}

	if err := s.otpService.SendOTP(user.Email, otpCode.Code); err != nil {
		return fmt.Errorf("failed to send registration otp: %w", err)
	}

	return nil
}

// VerifyRegistrationOTP verifies the user's registration activation code and activates their account status upon success.
func (s *Service) VerifyRegistrationOTP(ctx context.Context, req otp.VerifyOTPRequest) error {
	user, err := s.userRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return ErrInvalidCredentials
	}

	if err := s.otpService.VerifyOTP(ctx, user.ID, req.Code, otp.TypeRegistration); err != nil {
		return err
	}

	if err := s.userRepo.ActivateUser(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to activate account: %w", err)
	}

	return nil
}

// Login validates user credentials, checks activation state, and generates an RS256 signed JWT token upon success.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*User, string, error) {
	user, err := s.userRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, "", ErrAccountInactive
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", fmt.Errorf("failed to issue jwt token: %w", err)
	}

	return user, token, nil
}

// ForgotPassword initiates a password recovery flow by issuing an OTP if the account exists,
// while shielding user presence to prevent account enumeration attacks.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil
	}

	otpCode, err := s.otpService.GenerateOTP(ctx, user.ID, otp.TypePasswordReset)
	if err != nil {
		return fmt.Errorf("failed to generate password reset otp: %w", err)
	}

	if err := s.otpService.SendOTP(user.Email, otpCode.Code); err != nil {
		return fmt.Errorf("failed to send password reset otp: %w", err)
	}

	return nil
}

// ResetPassword validates the recovery OTP, hashes the new password, updates storage, and invalidates active OTP codes.
func (s *Service) ResetPassword(ctx context.Context, req otp.ResetPasswordRequest) error {
	user, err := s.userRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return ErrInvalidCredentials
	}

	if err := s.otpService.VerifyOTP(ctx, user.ID, req.Code, otp.TypePasswordReset); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdateUserPassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	_ = s.otpService.InvalidateUserOTPs(ctx, user.ID, otp.TypePasswordReset)

	return nil
}

// GetUserByID retrieves user entity by ID.
func (s *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.userRepo.FindUserByID(ctx, id)
}
