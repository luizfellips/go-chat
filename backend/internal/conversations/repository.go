package conversations

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateDirect(ctx context.Context, userID, participantID uuid.UUID) (*Conversation, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]ListItem, error)
	IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
	GetOtherParticipantID(ctx context.Context, conversationID, userID uuid.UUID) (uuid.UUID, error)
}
