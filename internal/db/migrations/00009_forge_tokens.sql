-- +goose Up
-- A token saved for reuse when registering repos, one per forge and (for
-- Forgejo) instance. instance_url is '' for GitHub, so the unique constraint
-- holds without NULL special cases. Encrypted like repos.token_encrypted.
CREATE TABLE forge_tokens (
    id              TEXT PRIMARY KEY,
    forge           TEXT NOT NULL,
    instance_url    TEXT NOT NULL DEFAULT '',
    token_encrypted BLOB NOT NULL,
    token_masked    TEXT NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    UNIQUE (forge, instance_url)
);

-- +goose Down
DROP TABLE forge_tokens;
