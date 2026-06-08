package messages

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/httpx"
	"github.com/luizf/go-chat/backend/internal/requestctx"
)

type Handler struct {
	svc *Service
	hub RealtimeHub
}

func NewHandler(svc *Service, hub RealtimeHub) *Handler {
	return &Handler{svc: svc, hub: hub}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestctx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, apperr.ErrUnauthorized)
		return
	}
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	var cursor *time.Time
	if c := r.URL.Query().Get("cursor"); c != "" {
		t, err := time.Parse(time.RFC3339Nano, c)
		if err != nil {
			t, err = time.Parse(time.RFC3339, c)
		}
		if err != nil {
			httpx.WriteError(w, apperr.ErrInvalidInput)
			return
		}
		cursor = &t
	}
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	out, err := h.svc.List(r.Context(), ListInput{
		ConversationID: convID,
		UserID:         userID,
		Cursor:         cursor,
		Limit:          int32(limit),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	type msgResp struct {
		ID        string  `json:"id"`
		SenderID  *string `json:"sender_id"`
		Content   string  `json:"content"`
		CreatedAt string  `json:"created_at"`
		ReadAt    *string `json:"read_at"`
	}
	msgs := make([]msgResp, 0, len(out.Messages))
	for _, m := range out.Messages {
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
		msgs = append(msgs, msgResp{
			ID: m.ID.String(), SenderID: senderID, Content: m.Content,
			CreatedAt: m.CreatedAt.Format(time.RFC3339Nano), ReadAt: readAt,
		})
	}
	resp := map[string]interface{}{"messages": msgs}
	if out.NextCursor != nil {
		resp["next_cursor"] = out.NextCursor.Format(time.RFC3339Nano)
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

type sendRequest struct {
	Content string `json:"content"`
}

func (r sendRequest) Valid(_ context.Context) map[string]string {
	return contentProblems(r.Content)
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestctx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, apperr.ErrUnauthorized)
		return
	}
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	req, problems, err := httpx.DecodeValid[sendRequest](r)
	if err != nil {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	if len(problems) > 0 {
		httpx.WriteValidationError(w, problems)
		return
	}
	msg, err := h.svc.Send(r.Context(), SendInput{
		ConversationID: convID, SenderID: userID, Content: req.Content,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	h.hub.BroadcastMessageReceived(r.Context(), msg)
	httpx.WriteJSON(w, http.StatusCreated, ToDTO(msg))
}
