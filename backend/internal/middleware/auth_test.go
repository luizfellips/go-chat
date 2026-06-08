package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/auth"
)

func TestRequireAuthValidToken(t *testing.T) {
	userID := uuid.New()
	svc := auth.NewTokenService("access-secret-key-32-chars-min!!", 15*time.Minute, time.Hour)
	token, _, err := svc.GenerateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}

	mw := NewAuthMiddleware(svc)
	called := false
	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}

func TestRequireAuthMissingToken(t *testing.T) {
	svc := auth.NewTokenService("access-secret-key-32-chars-min!!", 15*time.Minute, time.Hour)
	mw := NewAuthMiddleware(svc)
	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAuthQueryTokenNotAccepted(t *testing.T) {
	userID := uuid.New()
	svc := auth.NewTokenService("access-secret-key-32-chars-min!!", 15*time.Minute, time.Hour)
	token, _, err := svc.GenerateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}

	mw := NewAuthMiddleware(svc)
	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/?token="+token, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
