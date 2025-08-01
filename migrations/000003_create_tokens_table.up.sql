CREATE TABLE user_tokens (
    token VARCHAR(255) PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_user_tokens_expires_at ON user_tokens (expires_at);