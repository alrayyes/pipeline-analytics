-- +goose Up
ALTER TABLE webauthn_credentials ADD COLUMN label TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE webauthn_credentials DROP COLUMN label;
