package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"store_auth/prisma/db"
)

// ErrRefreshTokenNotFound indicates that the requested refresh token does not exist in persistence.
var ErrRefreshTokenNotFound = errors.New("refresh token not found")

// RefreshRepository abstracts database persistence operations for refresh tokens using Prisma Client Go.
type RefreshRepository struct {
	client *db.PrismaClient
}

// NewRefreshRepository constructs a RefreshRepository instance attached to an active Prisma client.
// Why: Provides dependency injection of database persistence for refresh token lifecycle management.
func NewRefreshRepository(client *db.PrismaClient) *RefreshRepository {
	return &RefreshRepository{
		client: client,
	}
}

// CreateRefreshToken persists a new hashed refresh token with user association and expiration timestamp.
// Why: Stores hashed credentials securely in PostgreSQL so only the client holds the raw opaque secret.
func (r *RefreshRepository) CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*RefreshToken, error) {
	created, err := r.client.RefreshToken.CreateOne(
		db.RefreshToken.User.Link(db.User.ID.Equals(userID)),
		db.RefreshToken.TokenHash.Set(tokenHash),
		db.RefreshToken.ExpiresAt.Set(expiresAt),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return mapPrismaRefreshTokenToModel(created), nil
}

// FindRefreshTokenByHash fetches a refresh token record matching the SHA-256 hash.
// Why: Enables lookup and revocation status verification during token rotation flows.
func (r *RefreshRepository) FindRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	token, err := r.client.RefreshToken.FindUnique(
		db.RefreshToken.TokenHash.Equals(tokenHash),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token by hash: %w", err)
	}

	return mapPrismaRefreshTokenToModel(token), nil
}

// RevokeRefreshToken marks a single refresh token record as revoked by its primary key ID.
// Why: Invalidates a specific rotated refresh token so it cannot be used again.
func (r *RefreshRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	_, err := r.client.RefreshToken.FindUnique(
		db.RefreshToken.ID.Equals(tokenID),
	).Update(
		db.RefreshToken.Revoked.Set(true),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrRefreshTokenNotFound
		}
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserRefreshTokens marks all active refresh tokens for a user as revoked.
// Why: Implements session invalidation when token reuse is detected or account security state changes.
func (r *RefreshRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.client.RefreshToken.FindMany(
		db.RefreshToken.UserID.Equals(userID),
		db.RefreshToken.Revoked.Equals(false),
	).Update(
		db.RefreshToken.Revoked.Set(true),
	).Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke all user refresh tokens: %w", err)
	}

	return nil
}

func mapPrismaRefreshTokenToModel(rt *db.RefreshTokenModel) *RefreshToken {
	return &RefreshToken{
		ID:        rt.ID,
		UserID:    rt.UserID,
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		Revoked:   rt.Revoked,
		CreatedAt: rt.CreatedAt,
	}
}
