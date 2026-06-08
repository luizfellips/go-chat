package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type envelope struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp string          `json:"timestamp"`
}

type Bot struct {
	id             int
	email          string
	username       string
	password       string
	api            *APIClient
	wsURL          string
	peerID         string
	conversationID string
	interval       time.Duration
	stats          *Stats
}

func (b *Bot) Run(ctx context.Context) {
	token, convID, err := b.prepare()
	if err != nil {
		b.stats.errors.Add(1)
		fmt.Printf("bot %d prepare failed: %v\n", b.id, err)
		return
	}
	b.conversationID = convID

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	ticket, err := b.api.GetWSTicket(token)
	if err != nil {
		b.stats.errors.Add(1)
		fmt.Printf("bot %d ws ticket failed: %v\n", b.id, err)
		return
	}
	conn, _, err := dialer.Dial(fmt.Sprintf("%s?ticket=%s", b.wsURL, ticket), http.Header{})
	if err != nil {
		b.stats.errors.Add(1)
		fmt.Printf("bot %d ws dial failed: %v\n", b.id, err)
		return
	}
	defer conn.Close()

	b.stats.connected.Add(1)
	defer b.stats.connected.Add(-1)

	readDone := make(chan struct{})
	go b.readLoop(conn, readDone)
	defer func() {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		<-readDone
	}()

	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	seq := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-readDone:
			b.stats.errors.Add(1)
			return
		case <-ticker.C:
			seq++
			if err := b.sendMessage(conn, seq); err != nil {
				b.stats.errors.Add(1)
				return
			}
			b.stats.sent.Add(1)
		}
	}
}

func (b *Bot) prepare() (token, conversationID string, err error) {
	if err := b.api.Register(b.email, b.username, b.password); err != nil {
		return "", "", err
	}

	token, err = b.api.Login(b.email, b.password)
	if err != nil {
		return "", "", err
	}

	conversationID, err = b.api.CreateConversation(token, b.peerID)
	if err != nil {
		return "", "", err
	}

	return token, conversationID, nil
}

func (b *Bot) sendMessage(conn *websocket.Conn, seq int) error {
	content := fmt.Sprintf("bot %d msg %d rand %d", b.id, seq, rand.Intn(10000))
	out := envelope{
		Type: "message_sent",
		Payload: mustJSON(map[string]string{
			"conversation_id": b.conversationID,
			"content":         content,
		}),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
	return conn.WriteJSON(out)
}

func (b *Bot) readLoop(conn *websocket.Conn, done chan struct{}) {
	defer close(done)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg envelope
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if msg.Type == "message_received" {
			b.stats.received.Add(1)
		}
	}
}

func mustJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
