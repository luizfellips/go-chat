package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luizf/go-chat/backend/internal/apperr"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) Create(ctx context.Context, email, username, passwordHash string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, password_hash, avatar_url, created_at, updated_at
	`, email, username, passwordHash)
	return scanUser(row)
}

func (r *PostgresRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, username, password_hash, avatar_url, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	return u, err
}

func (r *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, username, password_hash, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	return u, err
}

func (r *PostgresRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, username, password_hash, avatar_url, created_at, updated_at
		FROM users WHERE username = $1
	`, username)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	return u, err
}

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
