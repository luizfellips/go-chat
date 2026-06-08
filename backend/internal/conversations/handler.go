package conversations

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/httpx"
	"github.com/luizf/go-chat/backend/internal/requestctx"
)

type Handler struct {
	svc      *Service
	presence PresenceChecker
}

func NewHandler(svc *Service, presence PresenceChecker) *Handler {
	return &Handler{svc: svc, presence: presence}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestctx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, apperr.ErrUnauthorized)
		return
	}
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	type lastMsg struct {
		Content   *string `json:"content"`
		CreatedAt *string `json:"created_at"`
		SenderID  *string `json:"sender_id"`
	}
	type item struct {
		ID          string `json:"id"`
		Participant struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			IsOnline bool   `json:"is_online"`
		} `json:"participant"`
		LastMessage *lastMsg `json:"last_message"`
		UnreadCount int      `json:"unread_count"`
	}
	resp := make([]item, 0, len(items))
	for _, c := range items {
		var i item
		i.ID = c.ID.String()
		i.Participant.ID = c.ParticipantID.String()
		i.Participant.Username = c.ParticipantUsername
		i.Participant.IsOnline = h.presence.IsOnline(c.ParticipantID)
		i.UnreadCount = c.UnreadCount
		if c.LastMessagePreview != nil && c.LastMessageAt != nil {
			ts := c.LastMessageAt.Format(time.RFC3339Nano)
			var senderID *string
			if c.LastMessageSenderID != nil {
				s := c.LastMessageSenderID.String()
				senderID = &s
			}
			i.LastMessage = &lastMsg{Content: c.LastMessagePreview, CreatedAt: &ts, SenderID: senderID}
		}
		resp = append(resp, i)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{"conversations": resp})
}

type createRequest struct {
	ParticipantID string `json:"participant_id"`
}

func (r createRequest) Valid(_ context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.ParticipantID) == "" {
		problems["participant_id"] = "required"
		return problems
	}
	if _, err := uuid.Parse(r.ParticipantID); err != nil {
		problems["participant_id"] = "invalid format"
	}
	return problems
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestctx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, apperr.ErrUnauthorized)
		return
	}
	req, problems, err := httpx.DecodeValid[createRequest](r)
	if err != nil {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	if len(problems) > 0 {
		httpx.WriteValidationError(w, problems)
		return
	}
	participantID, _ := uuid.Parse(req.ParticipantID)
	conv, err := h.svc.Create(r.Context(), userID, participantID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":   conv.ID.String(),
		"type": conv.Type,
	})
}
