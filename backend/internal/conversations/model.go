package conversations

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID                 uuid.UUID
	Type               string
	LastMessageAt      *time.Time
	LastMessagePreview *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ListItem struct {
	Conversation
	ParticipantID        uuid.UUID
	ParticipantUsername  string
	ParticipantAvatarURL *string
	LastMessageSenderID  *uuid.UUID
	UnreadCount          int
}
