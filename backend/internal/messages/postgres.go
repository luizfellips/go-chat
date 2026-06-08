package messages

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luizf/go-chat/backend/internal/apperr"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) Create(ctx context.Context, conversationID, senderID uuid.UUID, content string) (*Message, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, sender_id, content, created_at, read_at
	`, conversationID, senderID, content)

	var msg Message
	if err := row.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.CreatedAt, &msg.ReadAt); err != nil {
		return nil, err
	}

	preview := content
	if len(preview) > 100 {
		preview = preview[:100]
	}
	if _, err := tx.Exec(ctx, `
		UPDATE conversations SET last_message_at = $2, last_message_preview = $3, updated_at = now() WHERE id = $1
	`, conversationID, msg.CreatedAt, preview); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *PostgresRepo) List(ctx context.Context, conversationID uuid.UUID, cursor *time.Time, limit int32) ([]Message, error) {
	var rows pgx.Rows
	var err error
	if cursor == nil {
		rows, err = r.pool.Query(ctx, `
			SELECT id, conversation_id, sender_id, content, created_at, read_at
			FROM messages WHERE conversation_id = $1
			ORDER BY created_at DESC LIMIT $2
		`, conversationID, limit)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, conversation_id, sender_id, content, created_at, read_at
			FROM messages WHERE conversation_id = $1 AND created_at < $2
			ORDER BY created_at DESC LIMIT $3
		`, conversationID, *cursor, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.CreatedAt, &m.ReadAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (r *PostgresRepo) MarkRead(ctx context.Context, messageID, conversationID, readerID uuid.UUID) (*Message, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE messages SET read_at = now()
		WHERE id = $1 AND conversation_id = $2 AND sender_id != $3 AND read_at IS NULL
		RETURNING id, conversation_id, sender_id, content, created_at, read_at
	`, messageID, conversationID, readerID)
	var m Message
	err := row.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.CreatedAt, &m.ReadAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	return &m, err
}

func (r *PostgresRepo) UpdateConversationLastMessage(ctx context.Context, conversationID uuid.UUID, preview string, at time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE conversations SET last_message_at = $2, last_message_preview = $3, updated_at = now() WHERE id = $1
	`, conversationID, at, preview)
	return err
}
