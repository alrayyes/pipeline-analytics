-- +goose Up
-- The table's inline UNIQUE (forge, forgejo_instance_url, identifier)
-- never actually stops a duplicate GitHub repo: forgejo_instance_url is
-- NULL for every GitHub row, and SQL treats every NULL as distinct from
-- every other NULL in a UNIQUE constraint, so two rows with identical
-- forge/identifier and both NULL instance URLs don't collide. Confirmed
-- live: inserting the same (forge, identifier) twice with a NULL
-- instance_url raises no error. COALESCE-ing the instance URL to '' in
-- an expression index makes NULL compare equal to itself for uniqueness
-- purposes, closing that gap without touching the column's nullability
-- (Forgejo rows still need a real instance URL; GitHub rows still don't
-- have one).
CREATE UNIQUE INDEX idx_repos_forge_identifier_instance
    ON repos (forge, identifier, COALESCE(forgejo_instance_url, ''));

-- +goose Down
DROP INDEX idx_repos_forge_identifier_instance;
