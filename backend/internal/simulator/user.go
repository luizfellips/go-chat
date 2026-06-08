package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type User struct {
	Index    int
	Email    string
	Username string
	Password string
	UserID   string
	Token    string
	ConvID   string
	PeerID   string
}

type wsEnvelope struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp string          `json:"timestamp"`
}

type messageReceivedPayload struct {
	Message struct {
		SenderID *string `json:"sender_id"`
		Content  string  `json:"content"`
	} `json:"message"`
}

func (u *User) Connect(ctx context.Context, wsURL string, api *APIClient, metrics *Metrics) (*websocket.Conn, error) {
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	ticket, err := api.GetWSTicket(u.Token)
	if err != nil {
		return nil, err
	}
	conn, _, err := dialer.Dial(fmt.Sprintf("%s?ticket=%s", wsURL, ticket), http.Header{})
	if err != nil {
		return nil, err
	}

	metrics.ConnUp()
	go u.readLoop(ctx, conn, metrics)

	return conn, nil
}

func (u *User) readLoop(ctx context.Context, conn *websocket.Conn, metrics *Metrics) {
	defer metrics.ConnDown()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var env wsEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		if env.Type != "message_received" {
			continue
		}

		var payload messageReceivedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			continue
		}
		if len(payload.Message.Content) >= len(latencyPrefix) && payload.Message.Content[:len(latencyPrefix)] == latencyPrefix {
			metrics.TrackReceive(payload.Message.Content, u.UserID)
		}
	}
}

func (u *User) SendMessage(conn *websocket.Conn, metrics *Metrics) error {
	key := latencyPrefix + uuid.NewString()
	env := wsEnvelope{
		Type: "message_sent",
		Payload: mustRawJSON(map[string]string{
			"conversation_id": u.ConvID,
			"content":         key,
		}),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}

	metrics.TrackSend(key, u.UserID)
	if err := conn.WriteJSON(env); err != nil {
		metrics.RecordError()
		return err
	}
	metrics.RecordSent()
	return nil
}

func mustRawJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
