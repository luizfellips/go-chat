package users

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, email, username, passwordHash string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
}
