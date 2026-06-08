package websocket

import (
	"time"

	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/messages"
)

const (
	EventConnection       = "connection"
	EventDisconnect       = "disconnect"
	EventMessageSent      = "message_sent"
	EventMessageReceived  = "message_received"
	EventTypingStart      = "typing_start"
	EventTypingStop       = "typing_stop"
	EventUserOnline       = "user_online"
	EventUserOffline      = "user_offline"
	EventMessageRead      = "message_read"
	EventError            = "error"
)

type Envelope struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp string      `json:"timestamp"`
}

func NewEnvelope(eventType string, payload interface{}) Envelope {
	return Envelope{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

type MessageSentPayload struct {
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

type MessageReceivedPayload struct {
	Message messages.DTO `json:"message"`
}

type TypingPayload struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
}

type UserPresencePayload struct {
	UserID string `json:"user_id"`
}

type MessageReadPayload struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	ReadAt         string `json:"read_at"`
}

type ConnectionPayload struct {
	UserID string `json:"user_id"`
}

type ErrorPayload struct {
	Code string `json:"code"`
}

func UserIDPayload(id uuid.UUID) UserPresencePayload {
	return UserPresencePayload{UserID: id.String()}
}
