package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type mockConversationAccess struct {
	participant bool
	err         error
}

func (m *mockConversationAccess) IsParticipant(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return m.participant, m.err
}

func TestHandleTypingRequiresParticipant(t *testing.T) {
	userID := uuid.New()
	otherID := uuid.New()
	convID := uuid.New()

	mock := &mockConversationAccess{participant: false}
	hub := NewHub(nil, mock, func(_ context.Context, conversationID, uid uuid.UUID) (uuid.UUID, error) {
		if conversationID == convID && uid == userID {
			return otherID, nil
		}
		return uuid.Nil, context.Canceled
	})

	client := &Client{
		hub:           hub,
		userID:        userID,
		ctx:           context.Background(),
		send:          make(chan []byte, 1),
		msgLimiter:    rate.NewLimiter(rate.Every(time.Minute/messageRatePerMin), messageRatePerMin),
		typingLimiter: rate.NewLimiter(rate.Every(time.Minute/typingRatePerMin), typingRatePerMin),
	}

	hub.handleTyping(client, NewEnvelope(EventTypingStart, TypingPayload{
		ConversationID: convID.String(),
	}))

	select {
	case msg := <-client.send:
		if string(msg) == "" {
			t.Fatal("expected error envelope")
		}
	default:
		t.Fatal("expected forbidden error to be sent to client")
	}
}
