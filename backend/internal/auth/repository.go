package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*RefreshToken, error)
	GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAndGetUserID(ctx context.Context, tokenHash string) (uuid.UUID, error)
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	RotateRefreshToken(ctx context.Context, oldHash, newHash string, expiresAt time.Time) (uuid.UUID, error)
}
