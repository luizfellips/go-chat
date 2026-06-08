-- +goose Up
CREATE TABLE direct_conversation_keys (
    user_low UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_high UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    PRIMARY KEY (user_low, user_high),
    CHECK (user_low < user_high)
);

CREATE UNIQUE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);

-- +goose Down
DROP INDEX IF EXISTS idx_refresh_tokens_hash;
DROP TABLE IF EXISTS direct_conversation_keys;
