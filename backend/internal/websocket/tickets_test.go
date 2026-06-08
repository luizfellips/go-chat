package websocket

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTicketStoreIssueAndRedeem(t *testing.T) {
	store := NewTicketStore(time.Minute)
	defer store.Stop()

	userID := uuid.New()
	ticket, err := store.Issue(userID)
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.Redeem(ticket)
	if err != nil {
		t.Fatal(err)
	}
	if got != userID {
		t.Fatalf("expected %v got %v", userID, got)
	}
}

func TestTicketStoreDoubleRedeemFails(t *testing.T) {
	store := NewTicketStore(time.Minute)
	defer store.Stop()

	ticket, err := store.Issue(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Redeem(ticket); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Redeem(ticket); err != ErrTicketInvalid {
		t.Fatalf("expected ErrTicketInvalid, got %v", err)
	}
}

func TestTicketStoreExpired(t *testing.T) {
	store := NewTicketStore(time.Millisecond)
	defer store.Stop()

	ticket, err := store.Issue(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err := store.Redeem(ticket); err != ErrTicketInvalid {
		t.Fatalf("expected ErrTicketInvalid, got %v", err)
	}
}
