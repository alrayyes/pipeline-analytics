-- +goose Up
-- Deletes duplicate rows before the unique index below can be created.
-- Necessary because the very bug that index closes (see its own comment)
-- could already have let duplicates in before #199 shipped -- confirmed
-- live: a database with pre-existing duplicates failed this migration
-- with "UNIQUE constraint failed: index
-- 'idx_repos_forge_identifier_instance'" before this step was added
-- (#247). Keeps the oldest row per (forge, identifier, instance) group
-- (earliest created_at, id as the tie-break) and drops the rest, which
-- cascades to that duplicate's own runs/jobs/steps -- the original
-- registration is the one worth keeping, not whichever duplicate
-- happened to register last.
DELETE FROM repos
WHERE id NOT IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY forge, identifier, COALESCE(forgejo_instance_url, '')
            ORDER BY created_at, id
        ) AS rn
        FROM repos
    )
    WHERE rn = 1
);

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
