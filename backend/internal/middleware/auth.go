package middleware

import (
	"net/http"
	"strings"

	"github.com/luizf/go-chat/backend/internal/auth"
	"github.com/luizf/go-chat/backend/internal/requestctx"
)

type AuthMiddleware struct {
	tokens *auth.TokenService
}

func NewAuthMiddleware(tokens *auth.TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		userID, err := m.tokens.ParseAccessToken(token)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := requestctx.WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}
