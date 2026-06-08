package server

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/time/rate"

	"github.com/luizf/go-chat/backend/internal/auth"
	"github.com/luizf/go-chat/backend/internal/config"
	"github.com/luizf/go-chat/backend/internal/conversations"
	"github.com/luizf/go-chat/backend/internal/health"
	"github.com/luizf/go-chat/backend/internal/messages"
	"github.com/luizf/go-chat/backend/internal/middleware"
	"github.com/luizf/go-chat/backend/internal/users"
	ws "github.com/luizf/go-chat/backend/internal/websocket"
)

type Handlers struct {
	Auth          *auth.Handler
	Users         *users.Handler
	Conversations *conversations.Handler
	Messages      *messages.Handler
	Health        *health.Handler
	WebSocket     *ws.Handler
}

func NewRouter(cfg config.Config, h Handlers, tokenService *auth.TokenService) chi.Router {
	authMW := middleware.NewAuthMiddleware(tokenService)

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/live", h.Health.Live)
	r.Head("/live", h.Health.Live)
	r.Get("/ready", h.Health.Ready)
	r.Head("/ready", h.Health.Ready)
	r.Get("/health", h.Health.Ready)
	r.Head("/health", h.Health.Ready)
	r.Handle("/metrics", middleware.MetricsAuth(cfg.MetricsToken, h.Health.Metrics()))
	r.Get("/ws/connect", h.WebSocket.Connect)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			if cfg.LoginRateLimitRPM > 0 {
				loginLimiter := middleware.NewRateLimiter(
					rate.Every(time.Minute/time.Duration(cfg.LoginRateLimitRPM)),
					cfg.LoginRateLimitRPM,
				)
				r.With(loginLimiter.Middleware).Post("/login", h.Auth.Login)
			} else {
				r.Post("/login", h.Auth.Login)
			}
			if cfg.RegisterRateLimitRPM > 0 {
				registerLimiter := middleware.NewRateLimiter(
					rate.Every(time.Minute/time.Duration(cfg.RegisterRateLimitRPM)),
					cfg.RegisterRateLimitRPM,
				)
				r.With(registerLimiter.Middleware).Post("/register", h.Auth.Register)
			} else {
				r.Post("/register", h.Auth.Register)
			}
			if cfg.RefreshRateLimitRPM > 0 {
				refreshLimiter := middleware.NewRateLimiter(
					rate.Every(time.Minute/time.Duration(cfg.RefreshRateLimitRPM)),
					cfg.RefreshRateLimitRPM,
				)
				r.With(refreshLimiter.Middleware).Post("/refresh", h.Auth.Refresh)
			} else {
				r.Post("/refresh", h.Auth.Refresh)
			}
			r.Post("/logout", h.Auth.Logout)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)
			r.Post("/ws/ticket", h.WebSocket.IssueTicket)
			r.Get("/users/me", h.Users.Me)
			r.Get("/users/search", h.Users.SearchByUsername)
			r.Get("/conversations", h.Conversations.List)
			r.Post("/conversations", h.Conversations.Create)
			r.Get("/conversations/{id}/messages", h.Messages.List)
			r.Post("/conversations/{id}/messages", h.Messages.Send)
		})
	})

	return r
}
