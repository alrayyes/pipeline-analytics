-- +goose Up
CREATE TABLE repos (
    id                       TEXT PRIMARY KEY,
    forge                    TEXT NOT NULL CHECK (forge IN ('github', 'forgejo')),
    identifier               TEXT NOT NULL,
    forgejo_instance_url     TEXT,
    token_encrypted          BLOB NOT NULL,
    token_masked             TEXT NOT NULL,
    webhook_secret           TEXT NOT NULL,
    ingestion_status         TEXT NOT NULL CHECK (ingestion_status IN ('pending', 'active', 'degraded')),
    ingestion_status_reason  TEXT,
    created_at               TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (forge, forgejo_instance_url, identifier)
);

CREATE TABLE runs (
    id            TEXT PRIMARY KEY,
    repo_id       TEXT NOT NULL REFERENCES repos (id) ON DELETE CASCADE,
    forge_run_id  TEXT NOT NULL,
    pipeline_name TEXT NOT NULL,
    status        TEXT NOT NULL,
    conclusion    TEXT,
    started_at    TIMESTAMP,
    completed_at  TIMESTAMP,
    forge_url     TEXT NOT NULL,
    UNIQUE (repo_id, forge_run_id)
);

CREATE INDEX idx_runs_repo_pipeline ON runs (repo_id, pipeline_name, started_at);

CREATE TABLE jobs (
    id            TEXT PRIMARY KEY,
    run_id        TEXT NOT NULL REFERENCES runs (id) ON DELETE CASCADE,
    forge_job_id  TEXT NOT NULL,
    name          TEXT NOT NULL,
    status        TEXT NOT NULL,
    conclusion    TEXT,
    queued_at     TIMESTAMP,
    started_at    TIMESTAMP,
    completed_at  TIMESTAMP,
    forge_url     TEXT NOT NULL,
    UNIQUE (run_id, forge_job_id)
);

CREATE INDEX idx_jobs_run ON jobs (run_id);

CREATE TABLE steps (
    id           TEXT PRIMARY KEY,
    job_id       TEXT NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    number       INTEGER NOT NULL,
    name         TEXT NOT NULL,
    status       TEXT NOT NULL,
    conclusion   TEXT,
    started_at   TIMESTAMP,
    completed_at TIMESTAMP,
    UNIQUE (job_id, number)
);

CREATE INDEX idx_steps_job ON steps (job_id);

-- +goose Down
DROP TABLE steps;
DROP TABLE jobs;
DROP TABLE runs;
DROP TABLE repos;
