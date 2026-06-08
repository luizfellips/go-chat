package bootstrap

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luizf/go-chat/backend/internal/auth"
	"github.com/luizf/go-chat/backend/internal/config"
	"github.com/luizf/go-chat/backend/internal/conversations"
	"github.com/luizf/go-chat/backend/internal/database"
	"github.com/luizf/go-chat/backend/internal/health"
	"github.com/luizf/go-chat/backend/internal/messages"
	"github.com/luizf/go-chat/backend/internal/server"
	"github.com/luizf/go-chat/backend/internal/users"
	ws "github.com/luizf/go-chat/backend/internal/websocket"
)

type Dependencies struct {
	Pool    *pgxpool.Pool
	Hub     *ws.Hub
	Tickets *ws.TicketStore
	Router  chi.Router
}

func Wire(ctx context.Context, cfg config.Config) (*Dependencies, error) {
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	userRepo := users.NewPostgresRepo(pool)
	refreshRepo := auth.NewPostgresRepo(pool)
	convRepo := conversations.NewPostgresRepo(pool)
	msgRepo := messages.NewPostgresRepo(pool)

	tokenService := auth.NewTokenService(
		cfg.JWTAccessSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	authSvc := auth.NewService(userRepo, refreshRepo, tokenService)
	usersSvc := users.NewService(userRepo)
	convSvc := conversations.NewService(userRepo, convRepo)
	msgSvc := messages.NewService(convRepo, msgRepo)

	tickets := ws.NewTicketStore(cfg.WSTicketTTL)
	hub := ws.NewHub(msgSvc, convRepo, convRepo.GetOtherParticipantID)
	wsHandler := ws.NewHandler(hub, tickets, cfg.CORSOrigin)

	router := server.NewRouter(cfg, server.Handlers{
		Auth:          auth.NewHandler(authSvc),
		Users:         users.NewHandler(usersSvc),
		Conversations: conversations.NewHandler(convSvc, hub),
		Messages:      messages.NewHandler(msgSvc, hub),
		Health:        health.NewHandler(pool),
		WebSocket:     wsHandler,
	}, tokenService)

	return &Dependencies{
		Pool:    pool,
		Hub:     hub,
		Tickets: tickets,
		Router:  router,
	}, nil
}
