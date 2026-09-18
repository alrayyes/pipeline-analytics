-- +goose Up
CREATE TABLE api_tokens (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES webauthn_users (id) ON DELETE CASCADE,
    token_hash BLOB NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_api_tokens_user ON api_tokens (user_id);

-- +goose Down
DROP TABLE api_tokens;
