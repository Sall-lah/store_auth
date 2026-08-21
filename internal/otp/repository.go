package otp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"store_auth/prisma/db"
)

// ErrOTPNotFound indicates that no matching OTP record exists for the given user and flow type.
var ErrOTPNotFound = errors.New("otp record not found")

// Repository handles persistence operations for 2FA numeric verification codes.
type Repository struct {
	client *db.PrismaClient
}

// NewRepository constructs a Repository tied to Prisma database client.
func NewRepository(client *db.PrismaClient) *Repository {
	return &Repository{
		client: client,
	}
}

// CreateOTP persists a newly generated OTP code linked to a specific user and purpose.
// Why: Stores short-lived verification codes for 2FA and registration activation.
func (r *Repository) CreateOTP(ctx context.Context, userID, code string, otpType Type, expiresAt time.Time) (*OTPCode, error) {
	prismaType := db.OTPTypeRegistration
	if otpType == TypePasswordReset {
		prismaType = db.OTPTypePasswordReset
	}

	created, err := r.client.OTPCode.CreateOne(
		db.OTPCode.User.Link(db.User.ID.Equals(userID)),
		db.OTPCode.Type.Set(prismaType),
		db.OTPCode.ExpiresAt.Set(expiresAt),
		db.OTPCode.Code.Set(code),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create otp: %w", err)
	}

	return mapPrismaOTPToModel(created), nil
}

// FindLatestOTPByUserAndType retrieves the most recently generated active OTP code for a user and action type.
// Why: Retrieves the current active pending verification code for validation.
func (r *Repository) FindLatestOTPByUserAndType(ctx context.Context, userID string, otpType Type) (*OTPCode, error) {
	prismaType := db.OTPTypeRegistration
	if otpType == TypePasswordReset {
		prismaType = db.OTPTypePasswordReset
	}

	otps, err := r.client.OTPCode.FindMany(
		db.OTPCode.UserID.Equals(userID),
		db.OTPCode.Type.Equals(prismaType),
		db.OTPCode.Used.Equals(false),
	).OrderBy(
		db.OTPCode.CreatedAt.Order(db.SortOrderDesc),
	).Take(1).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query latest otp: %w", err)
	}

	if len(otps) == 0 {
		return nil, ErrOTPNotFound
	}

	return mapPrismaOTPToModel(&otps[0]), nil
}

// IncrementOTPAttempts increments the invalid attempt counter on an OTP record to enforce rate throttling on brute force guesses.
// Why: Enforces brute force protection by limiting retry attempts before locking the code.
func (r *Repository) IncrementOTPAttempts(ctx context.Context, id string) error {
	_, err := r.client.OTPCode.FindUnique(
		db.OTPCode.ID.Equals(id),
	).Update(
		db.OTPCode.Attempts.Increment(1),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrOTPNotFound
		}
		return fmt.Errorf("failed to increment otp attempts: %w", err)
	}

	return nil
}

// MarkOTPUsed marks an OTP code as consumed in persistence.
// Why: Transitions the OTP record to used=true to prevent replay attacks and deactivate the verification code.
func (r *Repository) MarkOTPUsed(ctx context.Context, id string) error {
	_, err := r.client.OTPCode.FindUnique(
		db.OTPCode.ID.Equals(id),
	).Update(
		db.OTPCode.Used.Set(true),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrOTPNotFound
		}
		return fmt.Errorf("failed to mark otp as used: %w", err)
	}

	return nil
}

// MarkOTPUsedAndScrubCode marks an OTP code as consumed in persistence.
// Why: Alias for MarkOTPUsed to retain backward compatibility with callers while enforcing consumed state.
func (r *Repository) MarkOTPUsedAndScrubCode(ctx context.Context, id string) error {
	return r.MarkOTPUsed(ctx, id)
}

// InvalidateOTPsByUserAndType invalidates all pending OTP codes for a given user and type.
// Why: Revokes prior pending verification attempts upon new code issuance or lifecycle completion.
func (r *Repository) InvalidateOTPsByUserAndType(ctx context.Context, userID string, otpType Type) error {
	prismaType := db.OTPTypeRegistration
	if otpType == TypePasswordReset {
		prismaType = db.OTPTypePasswordReset
	}

	_, err := r.client.OTPCode.FindMany(
		db.OTPCode.UserID.Equals(userID),
		db.OTPCode.Type.Equals(prismaType),
		db.OTPCode.Used.Equals(false),
	).Update(
		db.OTPCode.Used.Set(true),
	).Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to invalidate otps: %w", err)
	}

	return nil
}

func mapPrismaOTPToModel(o *db.OTPCodeModel) *OTPCode {
	code, _ := o.Code()
	return &OTPCode{
		ID:        o.ID,
		UserID:    o.UserID,
		Code:      code,
		Type:      Type(o.Type),
		Attempts:  o.Attempts,
		ExpiresAt: o.ExpiresAt,
		Used:      o.Used,
		CreatedAt: o.CreatedAt,
	}
}
