package messages

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/apperr"
)

type ListInput struct {
	ConversationID uuid.UUID
	UserID         uuid.UUID
	Cursor         *time.Time
	Limit          int32
}

type ListOutput struct {
	Messages   []Message
	NextCursor *time.Time
}

type SendInput struct {
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Content        string
}

type MarkReadInput struct {
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	ReaderID       uuid.UUID
}

type Service struct {
	conversations ConversationAccess
	messages      Repository
}

func NewService(conversations ConversationAccess, messages Repository) *Service {
	return &Service{conversations: conversations, messages: messages}
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListOutput, error) {
	ok, err := s.conversations.IsParticipant(ctx, input.ConversationID, input.UserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperr.ErrForbidden
	}
	limit := input.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	msgs, err := s.messages.List(ctx, input.ConversationID, input.Cursor, limit)
	if err != nil {
		return nil, err
	}
	var nextCursor *time.Time
	if len(msgs) == int(limit) {
		oldest := msgs[len(msgs)-1].CreatedAt
		nextCursor = &oldest
	}
	return &ListOutput{Messages: msgs, NextCursor: nextCursor}, nil
}

func (s *Service) Send(ctx context.Context, input SendInput) (*Message, error) {
	if len(contentProblems(input.Content)) > 0 {
		return nil, apperr.ErrInvalidInput
	}
	ok, err := s.conversations.IsParticipant(ctx, input.ConversationID, input.SenderID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperr.ErrForbidden
	}
	return s.messages.Create(ctx, input.ConversationID, input.SenderID, input.Content)
}

func (s *Service) MarkRead(ctx context.Context, input MarkReadInput) (*Message, error) {
	ok, err := s.conversations.IsParticipant(ctx, input.ConversationID, input.ReaderID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperr.ErrForbidden
	}
	return s.messages.MarkRead(ctx, input.MessageID, input.ConversationID, input.ReaderID)
}
