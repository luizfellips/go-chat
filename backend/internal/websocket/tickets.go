package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrTicketInvalid = errors.New("invalid or expired ticket")

type ticketEntry struct {
	userID    uuid.UUID
	expiresAt time.Time
}

type TicketStore struct {
	mu      sync.Mutex
	tickets map[string]ticketEntry
	ttl     time.Duration
	stop    chan struct{}
}

func NewTicketStore(ttl time.Duration) *TicketStore {
	s := &TicketStore{
		tickets: make(map[string]ticketEntry),
		ttl:     ttl,
		stop:    make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

func (s *TicketStore) Stop() {
	close(s.stop)
}

func (s *TicketStore) Issue(userID uuid.UUID) (string, error) {
	raw, err := randomToken()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[raw] = ticketEntry{
		userID:    userID,
		expiresAt: time.Now().Add(s.ttl),
	}
	return raw, nil
}

func (s *TicketStore) Redeem(raw string) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.tickets[raw]
	if !ok {
		return uuid.Nil, ErrTicketInvalid
	}
	delete(s.tickets, raw)
	if time.Now().After(entry.expiresAt) {
		return uuid.Nil, ErrTicketInvalid
	}
	return entry.userID, nil
}

func (s *TicketStore) TTL() time.Duration {
	return s.ttl
}

func (s *TicketStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

func (s *TicketStore) cleanup() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.tickets {
		if now.After(v.expiresAt) {
			delete(s.tickets, k)
		}
	}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
