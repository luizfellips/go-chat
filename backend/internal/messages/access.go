package messages

import (
	"context"

	"github.com/google/uuid"
)

type ConversationAccess interface {
	IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
}
