package conversations

import (
	"context"
	"errors"

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

func orderedPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

func (r *PostgresRepo) CreateDirect(ctx context.Context, userID, participantID uuid.UUID) (*Conversation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	low, high := orderedPair(userID, participantID)

	var existingID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT conversation_id FROM direct_conversation_keys
		WHERE user_low = $1 AND user_high = $2
	`, low, high).Scan(&existingID)
	if err == nil {
		conv, err := r.getConversationTx(ctx, tx, existingID)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return conv, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO conversations (type) VALUES ('direct')
		RETURNING id, type, last_message_at, last_message_preview, created_at, updated_at
	`)
	var conv Conversation
	if err := row.Scan(&conv.ID, &conv.Type, &conv.LastMessageAt, &conv.LastMessagePreview, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
		return nil, err
	}

	for _, uid := range []uuid.UUID{userID, participantID} {
		if _, err := tx.Exec(ctx, `
			INSERT INTO conversation_participants (conversation_id, user_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, conv.ID, uid); err != nil {
			return nil, err
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO direct_conversation_keys (user_low, user_high, conversation_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_low, user_high) DO NOTHING
	`, low, high, conv.ID); err != nil {
		return nil, err
	}

	var finalID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT conversation_id FROM direct_conversation_keys
		WHERE user_low = $1 AND user_high = $2
	`, low, high).Scan(&finalID)
	if err != nil {
		return nil, err
	}
	if finalID != conv.ID {
		convPtr, err := r.getConversationTx(ctx, tx, finalID)
		if err != nil {
			return nil, err
		}
		conv = *convPtr
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *PostgresRepo) getConversationTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*Conversation, error) {
	row := tx.QueryRow(ctx, `
		SELECT id, type, last_message_at, last_message_preview, created_at, updated_at
		FROM conversations WHERE id = $1
	`, id)
	var conv Conversation
	if err := row.Scan(&conv.ID, &conv.Type, &conv.LastMessageAt, &conv.LastMessagePreview, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return &conv, nil
}

func (r *PostgresRepo) ListForUser(ctx context.Context, userID uuid.UUID) ([]ListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id, c.type, c.last_message_at, c.last_message_preview, c.created_at, c.updated_at,
			u.id, u.username, u.avatar_url,
			(
				SELECT m.sender_id FROM messages m
				WHERE m.conversation_id = c.id
				ORDER BY m.created_at DESC
				LIMIT 1
			),
			COALESCE((
				SELECT COUNT(*)::int FROM messages m
				WHERE m.conversation_id = c.id AND m.sender_id != $1 AND m.read_at IS NULL
			), 0)
		FROM conversations c
		JOIN conversation_participants cp ON cp.conversation_id = c.id AND cp.user_id = $1
		JOIN conversation_participants cp2 ON cp2.conversation_id = c.id AND cp2.user_id != $1
		JOIN users u ON u.id = cp2.user_id
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var item ListItem
		if err := rows.Scan(
			&item.ID, &item.Type, &item.LastMessageAt, &item.LastMessagePreview,
			&item.CreatedAt, &item.UpdatedAt,
			&item.ParticipantID, &item.ParticipantUsername, &item.ParticipantAvatarURL,
			&item.LastMessageSenderID,
			&item.UnreadCount,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepo) IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM conversation_participants
			WHERE conversation_id = $1 AND user_id = $2
		)
	`, conversationID, userID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepo) GetOtherParticipantID(ctx context.Context, conversationID, userID uuid.UUID) (uuid.UUID, error) {
	var otherID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT user_id FROM conversation_participants
		WHERE conversation_id = $1 AND user_id != $2 LIMIT 1
	`, conversationID, userID).Scan(&otherID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, apperr.ErrNotFound
	}
	return otherID, err
}
