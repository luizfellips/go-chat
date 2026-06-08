package conversations

import (
	"context"

	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/users"
)

type Service struct {
	users         users.Repository
	conversations Repository
}

func NewService(users users.Repository, conversations Repository) *Service {
	return &Service{users: users, conversations: conversations}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]ListItem, error) {
	return s.conversations.ListForUser(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID, participantID uuid.UUID) (*Conversation, error) {
	if userID == participantID {
		return nil, apperr.ErrInvalidInput
	}
	if _, err := s.users.GetByID(ctx, participantID); err != nil {
		return nil, err
	}
	return s.conversations.CreateDirect(ctx, userID, participantID)
}
