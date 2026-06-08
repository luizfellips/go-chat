package websocket

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/luizf/go-chat/backend/internal/httpx"
	"github.com/luizf/go-chat/backend/internal/requestctx"
)

type Handler struct {
	hub          *Hub
	tickets      *TicketStore
	allowedOrigin string
	upgrader     websocket.Upgrader
}

func NewHandler(hub *Hub, tickets *TicketStore, allowedOrigin string) *Handler {
	h := &Handler{
		hub:           hub,
		tickets:       tickets,
		allowedOrigin: allowedOrigin,
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	return h
}

func (h *Handler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	return origin == h.allowedOrigin
}

func (h *Handler) IssueTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestctx.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	ticket, err := h.tickets.Issue(userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ticket":     ticket,
		"expires_in": int64(h.tickets.TTL().Seconds()),
	})
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	ticket := strings.TrimSpace(r.URL.Query().Get("ticket"))
	if ticket == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := h.tickets.Redeem(ticket)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	ServeClient(h.hub, conn, userID)
}
