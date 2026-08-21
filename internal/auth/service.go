package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"store_auth/internal/jwt"
	"store_auth/internal/otp"
	platformRedis "store_auth/internal/platform/redis"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrAccountInactive     = errors.New("user account is inactive, please verify registration OTP")
	ErrAccountAlreadyActive = errors.New("account is already verified, please log in")
	ErrEmailTaken          = errors.New("email address is already registered")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrRefreshTokenReused  = errors.New("refresh token reuse detected")
	ErrInvalidOTPType      = errors.New("invalid otp type, must be 'registration' or 'password_reset'")
)

// Service encapsulates authentication logic, credential processing, and feature integration.
type Service struct {
	userRepo          *Repository
	refreshRepo       *RefreshRepository
	otpService        *otp.Service
	jwtService        *jwt.Service
	rdb               *redis.Client
	bcryptCost        int
	refreshExpiryDays int
}

// NewService constructs a Service binding user repository, refresh repository, Redis client, JWT signer, and OTP manager.
// Why: Injects all data access, cryptographic, and caching dependencies required for user authentication workflows.
func NewService(
	userRepo *Repository,
	refreshRepo *RefreshRepository,
	otpService *otp.Service,
	jwtService *jwt.Service,
	rdb *redis.Client,
	bcryptCost int,
	refreshExpiryDays int,
) *Service {
	if refreshExpiryDays <= 0 {
		refreshExpiryDays = 7
	}
	return &Service{
		userRepo:          userRepo,
		refreshRepo:       refreshRepo,
		otpService:        otpService,
		jwtService:        jwtService,
		rdb:               rdb,
		bcryptCost:        bcryptCost,
		refreshExpiryDays: refreshExpiryDays,
	}
}

// Register creates an inactive user account or refreshes credentials for an unverified account, generating and dispatching an OTP.
// Why: Initializes user identity in pending state until email ownership is proven, and prevents user lockout when re-registering an unverified account.
func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	existing, err := s.userRepo.FindUserByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		if existing.IsActive {
			return ErrEmailTaken
		}

		// Re-registration flow for inactive user: update name & password, invalidate old OTPs, issue new OTP
		user, err := s.userRepo.UpdateUnverifiedUserCredentials(ctx, existing.ID, req.Name, string(hashedPassword))
		if err != nil {
			return fmt.Errorf("failed to update unverified user credentials: %w", err)
		}

		_ = s.otpService.InvalidateUserOTPs(ctx, user.ID, otp.TypeRegistration)

		otpCode, err := s.otpService.GenerateOTP(ctx, user.ID, otp.TypeRegistration)
		if err != nil {
			return fmt.Errorf("failed to generate registration otp: %w", err)
		}

		if err := s.otpService.SendOTP(ctx, user.Email, otpCode.Code, user.Name, otp.TypeRegistration); err != nil {
			return fmt.Errorf("failed to send registration otp: %w", err)
		}

		return nil
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

	if err := s.otpService.SendOTP(ctx, user.Email, otpCode.Code, user.Name, otp.TypeRegistration); err != nil {
		return fmt.Errorf("failed to send registration otp: %w", err)
	}

	return nil
}

// VerifyRegistrationOTP verifies the user's registration activation code and activates their account status upon success.
// Why: Prevents unauthorized login until the owner validates the delivered one-time passcode and purges lingering verification codes.
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

	_ = s.otpService.InvalidateUserOTPs(ctx, user.ID, otp.TypeRegistration)

	return nil
}

// Login validates user credentials, checks account activation, and issues both access and refresh tokens.
// Why: Provides authenticated sessions with short-lived access credentials and rotatable refresh credentials.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*User, string, string, error) {
	user, err := s.userRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, "", "", ErrAccountInactive
	}

	accessToken, err := s.jwtService.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to issue jwt access token: %w", err)
	}

	refreshToken, err := s.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to issue refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// RefreshToken validates an incoming opaque refresh token, verifies user status, and rotates tokens.
// Why: Enables seamless session renewal without re-entering credentials while detecting compromised token replay.
func (s *Service) RefreshToken(ctx context.Context, rawToken string) (*User, string, string, error) {
	if rawToken == "" {
		return nil, "", "", ErrInvalidRefreshToken
	}

	tokenHash := hashToken(rawToken)
	rt, err := s.refreshRepo.FindRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return nil, "", "", ErrInvalidRefreshToken
		}
		return nil, "", "", err
	}

	// Reuse detection: if already revoked, revoke all tokens for this user as a security measure
	if rt.Revoked {
		_ = s.RevokeAllUserTokens(ctx, rt.UserID)
		return nil, "", "", ErrRefreshTokenReused
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, "", "", ErrRefreshTokenExpired
	}

	user, err := s.userRepo.FindUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, "", "", ErrInvalidRefreshToken
	}

	if !user.IsActive {
		_ = s.RevokeAllUserTokens(ctx, user.ID)
		return nil, "", "", ErrAccountInactive
	}

	// Rotate: Revoke the old token
	if err := s.refreshRepo.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return nil, "", "", fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Issue new refresh token
	newRefreshToken, err := s.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to issue rotated refresh token: %w", err)
	}

	// Issue new access token with latest user claims (e.g. updated role)
	newAccessToken, err := s.jwtService.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to issue new jwt access token: %w", err)
	}

	return user, newAccessToken, newRefreshToken, nil
}

// RevokeRefreshToken marks the specified refresh token as revoked upon user logout.
// Why: Ensures logged out refresh tokens cannot be used to generate future access tokens.
func (s *Service) RevokeRefreshToken(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	tokenHash := hashToken(rawToken)
	rt, err := s.refreshRepo.FindRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return nil
		}
		return err
	}

	if !rt.Revoked {
		return s.refreshRepo.RevokeRefreshToken(ctx, rt.ID)
	}
	return nil
}

// RevokeAllUserTokens revokes all database refresh tokens for a user and blacklists their ID in Redis.
// Why: Provides instantaneous invalidation of all existing access and refresh tokens when accounts are banned, deleted, or compromised.
func (s *Service) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if err := s.refreshRepo.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke user refresh tokens: %w", err)
	}

	if s.rdb != nil {
		if err := platformRedis.BlacklistUser(ctx, s.rdb, userID, 900*time.Second, "revoked_all_tokens"); err != nil {
			return fmt.Errorf("failed to blacklist user in redis: %w", err)
		}
	}

	return nil
}

// ForgotPassword initiates a password recovery flow by issuing an OTP if the account exists,
// while shielding user presence to prevent account enumeration attacks.
// Why: Safeguards account recovery without revealing registered email addresses to unauthenticated callers.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil
	}

	if !user.IsActive {
		return nil
	}

	_ = s.otpService.InvalidateUserOTPs(ctx, user.ID, otp.TypePasswordReset)

	otpCode, err := s.otpService.GenerateOTP(ctx, user.ID, otp.TypePasswordReset)
	if err != nil {
		return fmt.Errorf("failed to generate password reset otp: %w", err)
	}

	if err := s.otpService.SendOTP(ctx, user.Email, otpCode.Code, user.Name, otp.TypePasswordReset); err != nil {
		return fmt.Errorf("failed to send password reset otp: %w", err)
	}

	return nil
}

// ResendOTP dispatches a fresh OTP code for either registration activation or password recovery.
// Why: Enables users to request a new verification code if their initial code expired or was lost, with enumeration protection on reset flows.
func (s *Service) ResendOTP(ctx context.Context, req ResendOTPRequest) error {
	normalizedType := strings.ToLower(strings.TrimSpace(req.Type))
	switch normalizedType {
	case "registration":
		user, err := s.userRepo.FindUserByEmail(ctx, req.Email)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				return ErrUserNotFound
			}
			return err
		}

		if user.IsActive {
			return ErrAccountAlreadyActive
		}

		_ = s.otpService.InvalidateUserOTPs(ctx, user.ID, otp.TypeRegistration)

		otpCode, err := s.otpService.GenerateOTP(ctx, user.ID, otp.TypeRegistration)
		if err != nil {
			return fmt.Errorf("failed to generate registration otp: %w", err)
		}

		if err := s.otpService.SendOTP(ctx, user.Email, otpCode.Code, user.Name, otp.TypeRegistration); err != nil {
			return fmt.Errorf("failed to send registration otp: %w", err)
		}

		return nil

	case "password_reset":
		user, err := s.userRepo.FindUserByEmail(ctx, req.Email)
		if err != nil {
			// Anti-enumeration shield: return nil to avoid revealing user existence
			return nil
		}

		if !user.IsActive {
			// Do not send password reset OTPs to unverified accounts
			return nil
		}

		_ = s.otpService.InvalidateUserOTPs(ctx, user.ID, otp.TypePasswordReset)

		otpCode, err := s.otpService.GenerateOTP(ctx, user.ID, otp.TypePasswordReset)
		if err != nil {
			return fmt.Errorf("failed to generate password reset otp: %w", err)
		}

		if err := s.otpService.SendOTP(ctx, user.Email, otpCode.Code, user.Name, otp.TypePasswordReset); err != nil {
			return fmt.Errorf("failed to send password reset otp: %w", err)
		}

		return nil

	default:
		return ErrInvalidOTPType
	}
}

// ResetPassword validates the recovery OTP, hashes the new password, updates storage, and invalidates active OTP codes.
// Why: Ensures only verified OTP holders can update their password hash and prevents code reuse.
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
// Why: Provides internal service layer lookup for authenticated user profile retrieval.
func (s *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.userRepo.FindUserByID(ctx, id)
}

func (s *Service) issueRefreshToken(ctx context.Context, userID string) (string, error) {
	rawToken, err := generateRandomToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token bytes: %w", err)
	}

	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(time.Duration(s.refreshExpiryDays) * 24 * time.Hour)

	_, err = s.refreshRepo.CreateRefreshToken(ctx, userID, tokenHash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token in database: %w", err)
	}

	return rawToken, nil
}

func generateRandomToken(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

