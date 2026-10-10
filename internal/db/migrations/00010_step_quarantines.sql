-- +goose Up
-- A flaky step a person has marked as known, so it stops raising its
-- pipeline's flaky-step health signal. A step is one name within one pipeline.
-- A row is in force while expires_at is in the future; nothing sweeps expired
-- ones (they are ignored on read and replaced on the next mark), so health
-- stays recomputed per request.
CREATE TABLE step_quarantines (
    repo_id        TEXT NOT NULL REFERENCES repos (id) ON DELETE CASCADE,
    pipeline_name  TEXT NOT NULL,
    step_name      TEXT NOT NULL,
    note           TEXT NOT NULL DEFAULT '',
    quarantined_at TIMESTAMP NOT NULL,
    expires_at     TIMESTAMP NOT NULL,
    PRIMARY KEY (repo_id, pipeline_name, step_name)
);

-- +goose Down
DROP TABLE step_quarantines;
