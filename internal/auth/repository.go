package auth

import (
	"context"
	"errors"
	"fmt"

	"store_auth/prisma/db"
)

// ErrUserNotFound indicates that the requested user entity does not exist in persistence.
var ErrUserNotFound = errors.New("user not found")

// ErrEmailAlreadyExists indicates that a registration attempt used an email already registered.
var ErrEmailAlreadyExists = errors.New("email already exists")

// Repository abstracts database persistence operations for user accounts using Prisma Client Go.
type Repository struct {
	client *db.PrismaClient
}

// NewRepository constructs a Repository instance attached to an active Prisma client.
func NewRepository(client *db.PrismaClient) *Repository {
	return &Repository{
		client: client,
	}
}

// CreateUser persists a new pending user record into the database with isActive false.
func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, name string, role Role) (*User, error) {
	prismaRole := db.RoleCustomer
	if role == RoleAdmin {
		prismaRole = db.RoleAdmin
	}

	created, err := r.client.User.CreateOne(
		db.User.Email.Set(email),
		db.User.Password.Set(passwordHash),
		db.User.Name.Set(name),
		db.User.Role.Set(prismaRole),
	).Exec(ctx)

	if err != nil {
		if _, isUnique := db.IsErrUniqueConstraint(err); isUnique {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return mapPrismaUserToModel(created), nil
}

// FindUserByEmail fetches a user record matching the provided email address.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	u, err := r.client.User.FindUnique(
		db.User.Email.Equals(email),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return mapPrismaUserToModel(u), nil
}

// FindUserByID retrieves user profile data by primary key ID.
func (r *Repository) FindUserByID(ctx context.Context, id string) (*User, error) {
	u, err := r.client.User.FindUnique(
		db.User.ID.Equals(id),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return mapPrismaUserToModel(u), nil
}

// UpdateUserPassword updates the user's stored bcrypt password hash after successful password reset authorization.
func (r *Repository) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.client.User.FindUnique(
		db.User.ID.Equals(id),
	).Update(
		db.User.Password.Set(passwordHash),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}

// UpdateUnverifiedUserCredentials updates the user's display name and password hash during unverified account re-registration.
// Why: Allows pending inactive users to update their credentials or fix typos prior to confirming email ownership.
func (r *Repository) UpdateUnverifiedUserCredentials(ctx context.Context, id, name, passwordHash string) (*User, error) {
	updated, err := r.client.User.FindUnique(
		db.User.ID.Equals(id),
	).Update(
		db.User.Name.Set(name),
		db.User.Password.Set(passwordHash),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update unverified user credentials: %w", err)
	}

	return mapPrismaUserToModel(updated), nil
}

// ActivateUser sets is_active to true once registration OTP has been verified.
func (r *Repository) ActivateUser(ctx context.Context, id string) error {
	_, err := r.client.User.FindUnique(
		db.User.ID.Equals(id),
	).Update(
		db.User.IsActive.Set(true),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

// DeactivateUser sets is_active to false when a user is banned or suspended.
// Why: Enforces account deactivation in PostgreSQL so subsequent login and refresh attempts are rejected with 403 Forbidden.
func (r *Repository) DeactivateUser(ctx context.Context, id string) error {
	_, err := r.client.User.FindUnique(
		db.User.ID.Equals(id),
	).Update(
		db.User.IsActive.Set(false),
	).Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// DeleteUser removes a user record by primary key ID (cascading to OTP codes and refresh tokens).
// Why: Implements full account deletion compliance, freeing the unique email address and removing credentials.
func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.client.User.FindUnique(
		db.User.ID.Equals(id),
	).Delete().Exec(ctx)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func mapPrismaUserToModel(u *db.UserModel) *User {
	return &User{
		ID:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		Name:      u.Name,
		Role:      Role(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
