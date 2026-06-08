package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/auth"
)

func testTokenService(t *testing.T) *auth.TokenService {
	t.Helper()
	return auth.NewTokenService(
		"access-secret-key-32-chars-min!!",
		15*time.Minute,
		168*time.Hour,
	)
}

func TestCheckPassword(t *testing.T) {
	svc := testTokenService(t)

	hash, err := svc.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"correct password", "password123", true},
		{"wrong password", "wrong", false},
		{"empty password", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.CheckPassword(hash, tt.password); got != tt.want {
				t.Fatalf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"non-empty token", "test-token"},
		{"empty string", ""},
		{"uuid-like token", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := auth.HashToken(tt.input)
			b := auth.HashToken(tt.input)
			if a != b {
				t.Fatal("hash should be deterministic")
			}
			if a == "" {
				t.Fatal("hash should not be empty")
			}
		})
	}

	if auth.HashToken("a") == auth.HashToken("b") {
		t.Fatal("different inputs should produce different hashes")
	}
}

func TestParseAccessToken(t *testing.T) {
	svc := testTokenService(t)
	userID := uuid.New()

	validToken, _, err := svc.GenerateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}

	wrongSecretSvc := auth.NewTokenService(
		"other-secret-key-32-chars-min!!",
		15*time.Minute,
		168*time.Hour,
	)

	tests := []struct {
		name    string
		svc     *auth.TokenService
		token   string
		wantID  uuid.UUID
		wantErr bool
	}{
		{"valid token", svc, validToken, userID, false},
		{"malformed token", svc, "not-a-jwt", uuid.Nil, true},
		{"empty token", svc, "", uuid.Nil, true},
		{"wrong signing secret", wrongSecretSvc, validToken, uuid.Nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.svc.ParseAccessToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseAccessToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantID {
				t.Fatalf("ParseAccessToken() = %v, want %v", got, tt.wantID)
			}
		})
	}
}

func TestGenerateAccessTokenRoundTrip(t *testing.T) {
	svc := testTokenService(t)
	userID := uuid.New()

	token, expiresIn, err := svc.GenerateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected non-empty access token")
	}
	if expiresIn != int64((15 * time.Minute).Seconds()) {
		t.Fatalf("expiresIn = %d, want %d", expiresIn, int64((15*time.Minute).Seconds()))
	}

	parsed, err := svc.ParseAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed != userID {
		t.Fatalf("expected %v got %v", userID, parsed)
	}
}
