-- name: CreateConversation :one
INSERT INTO conversations (type) VALUES ('direct')
RETURNING id, type, last_message_at, last_message_preview, created_at, updated_at;

-- name: AddParticipant :exec
INSERT INTO conversation_participants (conversation_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: FindDirectConversation :one
SELECT c.id, c.type, c.last_message_at, c.last_message_preview, c.created_at, c.updated_at
FROM conversations c
JOIN conversation_participants cp1 ON cp1.conversation_id = c.id AND cp1.user_id = $1
JOIN conversation_participants cp2 ON cp2.conversation_id = c.id AND cp2.user_id = $2
WHERE c.type = 'direct'
LIMIT 1;

-- name: IsParticipant :one
SELECT EXISTS(
    SELECT 1 FROM conversation_participants
    WHERE conversation_id = $1 AND user_id = $2
) AS is_participant;

-- name: GetOtherParticipantID :one
SELECT user_id FROM conversation_participants
WHERE conversation_id = $1 AND user_id != $2
LIMIT 1;

-- name: ListConversationsForUser :many
SELECT
    c.id,
    c.type,
    c.last_message_at,
    c.last_message_preview,
    c.created_at,
    c.updated_at,
    u.id AS participant_id,
    u.username AS participant_username,
    u.avatar_url AS participant_avatar_url,
    COALESCE((
        SELECT COUNT(*)::int FROM messages m
        WHERE m.conversation_id = c.id
          AND m.sender_id != $1
          AND m.read_at IS NULL
    ), 0) AS unread_count
FROM conversations c
JOIN conversation_participants cp ON cp.conversation_id = c.id AND cp.user_id = $1
JOIN conversation_participants cp2 ON cp2.conversation_id = c.id AND cp2.user_id != $1
JOIN users u ON u.id = cp2.user_id
ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC;
