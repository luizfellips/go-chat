-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_id, content)
VALUES ($1, $2, $3)
RETURNING id, conversation_id, sender_id, content, created_at, read_at;

-- name: ListMessages :many
SELECT id, conversation_id, sender_id, content, created_at, read_at
FROM messages
WHERE conversation_id = $1
  AND (sqlc.narg('cursor')::timestamptz IS NULL OR created_at < sqlc.narg('cursor')::timestamptz)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: MarkMessageRead :one
UPDATE messages
SET read_at = now()
WHERE id = $1 AND conversation_id = $2 AND sender_id != $3 AND read_at IS NULL
RETURNING id, conversation_id, sender_id, content, created_at, read_at;

-- name: UpdateConversationLastMessage :exec
UPDATE conversations
SET last_message_at = $2, last_message_preview = $3, updated_at = now()
WHERE id = $1;

-- name: GetUnreadCount :one
SELECT COUNT(*)::int AS count FROM messages
WHERE conversation_id = $1 AND sender_id != $2 AND read_at IS NULL;
