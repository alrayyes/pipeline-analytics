-- +goose Up
ALTER TABLE runs ADD COLUMN branch TEXT;
ALTER TABLE runs ADD COLUMN head_sha TEXT;
ALTER TABLE runs ADD COLUMN head_message TEXT;
ALTER TABLE runs ADD COLUMN actor TEXT;

-- +goose Down
ALTER TABLE runs DROP COLUMN actor;
ALTER TABLE runs DROP COLUMN head_message;
ALTER TABLE runs DROP COLUMN head_sha;
ALTER TABLE runs DROP COLUMN branch;
