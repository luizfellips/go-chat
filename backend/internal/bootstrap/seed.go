package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/auth"
	"github.com/luizf/go-chat/backend/internal/config"
	"github.com/luizf/go-chat/backend/internal/database"
	"github.com/luizf/go-chat/backend/internal/users"
)

func RunSeed(cfg config.Config) error {
	if !cfg.SeedDemoUsers {
		log.Info().Msg("seed skipped (SEED_DEMO_USERS=false)")
		return nil
	}

	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer pool.Close()

	userRepo := users.NewPostgresRepo(pool)
	tokenService := auth.NewTokenService(
		cfg.JWTAccessSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	seedUsers := []struct{ email, username, password string }{
		{"alice@example.com", "alice", "password123"},
		{"bob@example.com", "bob", "password123"},
	}
	for _, u := range seedUsers {
		if _, err := userRepo.GetByEmail(ctx, u.email); err == nil {
			continue
		} else if !errors.Is(err, apperr.ErrNotFound) {
			log.Warn().Err(err).Str("email", u.email).Msg("seed lookup failed")
			continue
		}
		hash, _ := tokenService.HashPassword(u.password)
		if _, err := userRepo.Create(ctx, u.email, u.username, hash); err != nil {
			log.Warn().Err(err).Str("email", u.email).Msg("seed user failed")
		}
	}

	log.Info().Msg("seed completed")
	return nil
}
