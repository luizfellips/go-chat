package messages

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, conversationID, senderID uuid.UUID, content string) (*Message, error)
	List(ctx context.Context, conversationID uuid.UUID, cursor *time.Time, limit int32) ([]Message, error)
	MarkRead(ctx context.Context, messageID, conversationID, readerID uuid.UUID) (*Message, error)
	UpdateConversationLastMessage(ctx context.Context, conversationID uuid.UUID, preview string, at time.Time) error
}
