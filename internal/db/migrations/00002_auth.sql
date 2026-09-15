-- +goose Up
CREATE TABLE webauthn_users (
    id           TEXT PRIMARY KEY,
    user_handle  BLOB NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE webauthn_credentials (
    id         BLOB PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES webauthn_users (id) ON DELETE CASCADE,
    data       BLOB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_webauthn_credentials_user ON webauthn_credentials (user_id);

CREATE TABLE webauthn_ceremonies (
    id         TEXT PRIMARY KEY,
    data       BLOB NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE sessions (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES webauthn_users (id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user ON sessions (user_id);

-- +goose Down
DROP TABLE sessions;
DROP TABLE webauthn_ceremonies;
DROP TABLE webauthn_credentials;
DROP TABLE webauthn_users;
