-- +goose Up
CREATE TABLE account_settings (
    user_id    TEXT PRIMARY KEY REFERENCES webauthn_users (id) ON DELETE CASCADE,
    data       TEXT NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE account_settings;
