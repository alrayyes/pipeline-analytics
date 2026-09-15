-- +goose Up
ALTER TABLE repos ADD COLUMN reconcile_etag TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE repos DROP COLUMN reconcile_etag;
