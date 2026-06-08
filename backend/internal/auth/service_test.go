package auth

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/users"
)

type mockUserRepo struct {
	users map[uuid.UUID]*users.User
}

func (m *mockUserRepo) Create(_ context.Context, email, username, hash string) (*users.User, error) {
	id := uuid.New()
	u := &users.User{ID: id, Email: email, Username: username, PasswordHash: hash}
	m.users[id] = u
	return u, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*users.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, apperr.ErrNotFound
}

func (m *mockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*users.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*users.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, apperr.ErrNotFound
}

type mockRefreshRepo struct {
	mu     sync.Mutex
	tokens map[string]refreshEntry
}

type refreshEntry struct {
	userID    uuid.UUID
	revoked   bool
	expiresAt time.Time
}

func newMockRefreshRepo() *mockRefreshRepo {
	return &mockRefreshRepo{tokens: make(map[string]refreshEntry)}
}

func (m *mockRefreshRepo) Create(_ context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[tokenHash] = refreshEntry{userID: userID, expiresAt: expiresAt}
	return &RefreshToken{UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt}, nil
}

func (m *mockRefreshRepo) GetByHash(_ context.Context, tokenHash string) (*RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.tokens[tokenHash]
	if !ok || e.revoked || time.Now().After(e.expiresAt) {
		return nil, apperr.ErrNotFound
	}
	return &RefreshToken{UserID: e.userID, TokenHash: tokenHash, ExpiresAt: e.expiresAt}, nil
}

func (m *mockRefreshRepo) Revoke(_ context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.tokens[tokenHash]
	if !ok {
		return apperr.ErrNotFound
	}
	e.revoked = true
	m.tokens[tokenHash] = e
	return nil
}

func (m *mockRefreshRepo) RevokeAndGetUserID(_ context.Context, tokenHash string) (uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.tokens[tokenHash]
	if !ok || e.revoked || time.Now().After(e.expiresAt) {
		return uuid.Nil, apperr.ErrNotFound
	}
	e.revoked = true
	m.tokens[tokenHash] = e
	return e.userID, nil
}

func (m *mockRefreshRepo) RevokeAllForUser(_ context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for hash, e := range m.tokens {
		if e.userID == userID {
			e.revoked = true
			m.tokens[hash] = e
		}
	}
	return nil
}

func (m *mockRefreshRepo) RotateRefreshToken(_ context.Context, oldHash, newHash string, expiresAt time.Time) (uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.tokens[oldHash]
	if !ok || e.revoked || time.Now().After(e.expiresAt) {
		return uuid.Nil, apperr.ErrNotFound
	}
	e.revoked = true
	m.tokens[oldHash] = e
	m.tokens[newHash] = refreshEntry{userID: e.userID, expiresAt: expiresAt}
	return e.userID, nil
}

func TestRefreshConcurrentRotation(t *testing.T) {
	userID := uuid.New()
	userRepo := &mockUserRepo{users: map[uuid.UUID]*users.User{
		userID: {ID: userID, Email: "a@b.com", Username: "alice"},
	}}
	refreshRepo := newMockRefreshRepo()
	tokens := NewTokenService("access-secret-key-32-chars-min!!", 15*time.Minute, time.Hour)

	raw := uuid.New().String()
	hash := HashToken(raw)
	refreshRepo.tokens[hash] = refreshEntry{userID: userID, expiresAt: time.Now().Add(time.Hour)}

	svc := NewService(userRepo, refreshRepo, tokens)

	var wg sync.WaitGroup
	successes := make(chan struct{}, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.Refresh(context.Background(), raw); err == nil {
				successes <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(successes)

	count := 0
	for range successes {
		count++
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 successful refresh, got %d", count)
	}
}
