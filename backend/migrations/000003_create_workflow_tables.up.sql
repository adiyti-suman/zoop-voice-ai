CREATE TABLE IF NOT EXISTS workflows (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      VARCHAR(50)  NOT NULL DEFAULT 'DRAFT',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_versions (
    id          UUID PRIMARY KEY,
    workflow_id UUID         NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    version     VARCHAR(100) NOT NULL,
    status      VARCHAR(50)  NOT NULL DEFAULT 'DRAFT',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    archived_at  TIMESTAMPTZ,
    UNIQUE (workflow_id, version)
);

CREATE TABLE IF NOT EXISTS workflow_nodes (
    id                  UUID PRIMARY KEY,
    workflow_version_id UUID         NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE,
    node_key            VARCHAR(255) NOT NULL,
    type                VARCHAR(100) NOT NULL,
    name                VARCHAR(255) NOT NULL,
    config              JSONB        NOT NULL DEFAULT '{}',
    position            JSONB        NOT NULL DEFAULT '{"x":0,"y":0}',
    timeout_ms          INTEGER      NOT NULL DEFAULT 0,
    retry_policy        JSONB,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (workflow_version_id, node_key)
);

CREATE TABLE IF NOT EXISTS workflow_edges (
    id                  UUID PRIMARY KEY,
    workflow_version_id UUID         NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE,
    source_node_key     VARCHAR(255) NOT NULL,
    target_node_key     VARCHAR(255) NOT NULL,
    condition           TEXT         NOT NULL DEFAULT '',
    metadata            JSONB        NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_runs (
    id                  UUID PRIMARY KEY,
    verification_id     UUID         NOT NULL REFERENCES verifications(id),
    workflow_version_id UUID         NOT NULL REFERENCES workflow_versions(id),
    status              VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    current_node_key    VARCHAR(255),
    input               JSONB        NOT NULL DEFAULT '{}',
    output              JSONB        NOT NULL DEFAULT '{}',
    error               JSONB,
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_verification_id     ON workflow_runs(verification_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_version_id ON workflow_runs(workflow_version_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_status              ON workflow_runs(status);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_created_at          ON workflow_runs(created_at);

CREATE TABLE IF NOT EXISTS workflow_node_runs (
    id              UUID PRIMARY KEY,
    workflow_run_id UUID         NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    node_key        VARCHAR(255) NOT NULL,
    attempt         INTEGER      NOT NULL DEFAULT 1,
    status          VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    input           JSONB        NOT NULL DEFAULT '{}',
    output          JSONB        NOT NULL DEFAULT '{}',
    error           JSONB,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    duration_ms     BIGINT       NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (workflow_run_id, node_key, attempt)
);

CREATE INDEX IF NOT EXISTS idx_node_runs_workflow_run_id ON workflow_node_runs(workflow_run_id);
CREATE INDEX IF NOT EXISTS idx_node_runs_status          ON workflow_node_runs(status);
