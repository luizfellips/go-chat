//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/luizf/go-chat/backend/internal/bootstrap"
)

func integrationGetenv(databaseURL string) func(string) string {
	return func(key string) string {
		switch key {
		case "DATABASE_URL":
			return databaseURL
		case "PORT":
			return "0"
		case "APP_ENV":
			return "development"
		case "JWT_ACCESS_SECRET":
			return "dev-access-secret-change-in-production-32"
		case "SEED_DEMO_USERS":
			return "false"
		default:
			return ""
		}
	}
}

func TestRegisterLoginAndWSTicket(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	getenv := integrationGetenv(databaseURL)

	if err := bootstrap.Run(ctx, getenv, []string{"migrate"}); err != nil {
		t.Fatal(err)
	}

	srv, err := bootstrap.StartHTTP(ctx, getenv)
	if err != nil {
		t.Fatal(err)
	}

	email := "integration-" + time.Now().Format("150405") + "@example.com"
	registerBody, _ := json.Marshal(map[string]string{
		"email": email, "username": "intuser", "password": "password123",
	})
	res, err := http.Post(srv.BaseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(registerBody))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d", res.StatusCode)
	}
	res.Body.Close()

	loginBody, _ := json.Marshal(map[string]string{"email": email, "password": "password123"})
	res, err = http.Post(srv.BaseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d", res.StatusCode)
	}

	var loginResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.BaseURL+"/api/v1/ws/ticket", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("ws ticket status = %d", res.StatusCode)
	}
	var ticketResp struct {
		Ticket string `json:"ticket"`
	}
	if err := json.NewDecoder(res.Body).Decode(&ticketResp); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if ticketResp.Ticket == "" {
		t.Fatal("expected non-empty ticket")
	}
}
