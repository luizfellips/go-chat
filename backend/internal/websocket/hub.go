package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/messages"
	"github.com/luizf/go-chat/backend/internal/metrics"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 4096

	messageRatePerMin = 30
	typingRatePerMin  = 20
)

type ConversationAccess interface {
	IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
}

type Hub struct {
	clients    map[uuid.UUID]*Client
	online     map[uuid.UUID]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan targetEvent
	stop       chan struct{}
	done       chan struct{}
	mu         sync.RWMutex

	messages     *messages.Service
	conversations ConversationAccess
	getOther     func(ctx context.Context, conversationID, userID uuid.UUID) (uuid.UUID, error)
}

type targetEvent struct {
	userIDs []uuid.UUID
	data    []byte
	all     bool
}

func NewHub(
	messages *messages.Service,
	conversations ConversationAccess,
	getOther func(ctx context.Context, conversationID, userID uuid.UUID) (uuid.UUID, error),
) *Hub {
	return &Hub{
		clients:       make(map[uuid.UUID]*Client),
		online:        make(map[uuid.UUID]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan targetEvent, 256),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
		messages:      messages,
		conversations: conversations,
		getOther:      getOther,
	}
}

func (h *Hub) Run() {
	defer close(h.done)
	for {
		select {
		case <-h.stop:
			h.closeAllClients()
			return
		case client := <-h.register:
			h.mu.Lock()
			if old, ok := h.clients[client.userID]; ok {
				close(old.send)
				old.conn.Close()
			}
			h.clients[client.userID] = client
			h.online[client.userID] = true
			metrics.WSConnectionsActive.Set(float64(len(h.clients)))
			onlineSnapshot := make([]uuid.UUID, 0, len(h.online))
			for uid := range h.online {
				if uid != client.userID {
					onlineSnapshot = append(onlineSnapshot, uid)
				}
			}
			h.mu.Unlock()

			client.sendJSON(NewEnvelope(EventConnection, ConnectionPayload{UserID: client.userID.String()}))
			for _, uid := range onlineSnapshot {
				client.sendJSON(NewEnvelope(EventUserOnline, UserIDPayload(uid)))
			}
			h.broadcastAll(NewEnvelope(EventUserOnline, UserIDPayload(client.userID)))

		case client := <-h.unregister:
			h.mu.Lock()
			if c, ok := h.clients[client.userID]; ok && c == client {
				delete(h.clients, client.userID)
				delete(h.online, client.userID)
				close(client.send)
				metrics.WSConnectionsActive.Set(float64(len(h.clients)))
				h.mu.Unlock()
				h.broadcastAll(NewEnvelope(EventUserOffline, UserIDPayload(client.userID)))
			} else {
				h.mu.Unlock()
			}

		case evt := <-h.broadcast:
			h.mu.RLock()
			if evt.all {
				for _, c := range h.clients {
					select {
					case c.send <- evt.data:
					default:
						log.Warn().Str("user_id", c.userID.String()).Msg("client send buffer full")
					}
				}
			} else {
				for _, uid := range evt.userIDs {
					if c, ok := h.clients[uid]; ok {
						select {
						case c.send <- evt.data:
						default:
							log.Warn().Str("user_id", uid.String()).Msg("client send buffer full")
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.stop)
	<-h.done
}

func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, c := range h.clients {
		c.cancel()
		c.conn.Close()
	}
	h.clients = make(map[uuid.UUID]*Client)
	h.online = make(map[uuid.UUID]bool)
	metrics.WSConnectionsActive.Set(0)
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.online[userID]
}

func (h *Hub) broadcastAll(env Envelope) {
	data, _ := json.Marshal(env)
	h.broadcast <- targetEvent{all: true, data: data}
}

func (h *Hub) sendToUsers(userIDs []uuid.UUID, env Envelope) {
	data, _ := json.Marshal(env)
	h.broadcast <- targetEvent{userIDs: userIDs, data: data}
}

func (h *Hub) BroadcastMessageReceived(ctx context.Context, msg *messages.Message) {
	if msg.SenderID == nil {
		return
	}
	env := NewEnvelope(EventMessageReceived, MessageReceivedPayload{Message: messages.ToDTO(msg)})
	h.broadcastConversation(ctx, msg.ConversationID, *msg.SenderID, env)
}

func (h *Hub) broadcastConversation(ctx context.Context, conversationID, senderID uuid.UUID, env Envelope) {
	other, err := h.getOther(ctx, conversationID, senderID)
	if err != nil {
		return
	}
	h.sendToUsers([]uuid.UUID{senderID, other}, env)
}

func (h *Hub) HandleIncoming(client *Client, env Envelope) {
	switch env.Type {
	case EventMessageSent:
		h.handleMessageSent(client, env)
	case EventMessageRead:
		h.handleMessageRead(client, env)
	case EventTypingStart, EventTypingStop:
		h.handleTyping(client, env)
	}
}

func (h *Hub) handleMessageSent(client *Client, env Envelope) {
	if !client.msgLimiter.Allow() {
		client.sendError("rate_limit_exceeded")
		client.conn.Close()
		return
	}
	raw, _ := json.Marshal(env.Payload)
	var p MessageSentPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}
	convID, err := uuid.Parse(p.ConversationID)
	if err != nil {
		return
	}
	msg, err := h.messages.Send(client.ctx, messages.SendInput{
		ConversationID: convID,
		SenderID:       client.userID,
		Content:        p.Content,
	})
	if err != nil {
		if err == apperr.ErrForbidden {
			client.sendError("forbidden")
			return
		}
		log.Error().Err(err).Msg("ws send message failed")
		return
	}
	metrics.MessagesSentTotal.Inc()
	h.BroadcastMessageReceived(client.ctx, msg)
}

func (h *Hub) handleMessageRead(client *Client, env Envelope) {
	raw, _ := json.Marshal(env.Payload)
	var p MessageReadPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}
	convID, err := uuid.Parse(p.ConversationID)
	if err != nil {
		return
	}
	msgID, err := uuid.Parse(p.MessageID)
	if err != nil {
		return
	}
	msg, err := h.messages.MarkRead(client.ctx, messages.MarkReadInput{
		MessageID: msgID, ConversationID: convID, ReaderID: client.userID,
	})
	if err != nil {
		if err == apperr.ErrForbidden {
			client.sendError("forbidden")
		}
		return
	}
	readAt := msg.ReadAt.Format(time.RFC3339Nano)
	readEnv := NewEnvelope(EventMessageRead, MessageReadPayload{
		ConversationID: convID.String(),
		MessageID:      msg.ID.String(),
		ReadAt:         readAt,
	})
	if msg.SenderID != nil {
		h.sendToUsers([]uuid.UUID{*msg.SenderID, client.userID}, readEnv)
	}
}

func (h *Hub) handleTyping(client *Client, env Envelope) {
	if !client.typingLimiter.Allow() {
		return
	}
	raw, _ := json.Marshal(env.Payload)
	var p TypingPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}
	convID, err := uuid.Parse(p.ConversationID)
	if err != nil {
		return
	}
	ok, err := h.conversations.IsParticipant(client.ctx, convID, client.userID)
	if err != nil || !ok {
		client.sendError("forbidden")
		return
	}
	other, err := h.getOther(client.ctx, convID, client.userID)
	if err != nil {
		return
	}
	p.UserID = client.userID.String()
	env.Payload = p
	h.sendToUsers([]uuid.UUID{other}, env)
}

type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	userID        uuid.UUID
	send          chan []byte
	ctx           context.Context
	cancel        context.CancelFunc
	msgLimiter    *rate.Limiter
	typingLimiter *rate.Limiter
}

func (c *Client) sendJSON(env Envelope) {
	data, err := json.Marshal(env)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) sendError(code string) {
	c.sendJSON(NewEnvelope(EventError, ErrorPayload{Code: code}))
}

func (c *Client) readPump() {
	defer func() {
		c.cancel()
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		c.hub.HandleIncoming(c, env)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		hub:           hub,
		conn:          conn,
		userID:        userID,
		send:          make(chan []byte, 256),
		ctx:           ctx,
		cancel:        cancel,
		msgLimiter:    rate.NewLimiter(rate.Every(time.Minute/messageRatePerMin), messageRatePerMin),
		typingLimiter: rate.NewLimiter(rate.Every(time.Minute/typingRatePerMin), typingRatePerMin),
	}
	go client.writePump()
	go client.readPump()
	hub.Register(client)
}
