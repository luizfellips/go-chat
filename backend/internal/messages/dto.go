package messages

import "time"

type DTO struct {
	ID             string  `json:"id"`
	ConversationID string  `json:"conversation_id"`
	SenderID       *string `json:"sender_id"`
	Content        string  `json:"content"`
	CreatedAt      string  `json:"created_at"`
	ReadAt         *string `json:"read_at,omitempty"`
}

func ToDTO(m *Message) DTO {
	var senderID *string
	if m.SenderID != nil {
		s := m.SenderID.String()
		senderID = &s
	}
	var readAt *string
	if m.ReadAt != nil {
		ra := m.ReadAt.Format(time.RFC3339Nano)
		readAt = &ra
	}
	return DTO{
		ID:             m.ID.String(),
		ConversationID: m.ConversationID.String(),
		SenderID:       senderID,
		Content:        m.Content,
		CreatedAt:      m.CreatedAt.Format(time.RFC3339Nano),
		ReadAt:         readAt,
	}
}
