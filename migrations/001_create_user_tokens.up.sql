CREATE TABLE user_tokens (
    username TEXT PRIMARY KEY NOT NULL,
    token TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_tokens_created_at ON user_tokens(created_at);