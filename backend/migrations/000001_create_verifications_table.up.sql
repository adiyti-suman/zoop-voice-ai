CREATE TABLE IF NOT EXISTS verifications (
    id UUID PRIMARY KEY,
    external_id VARCHAR(255),
    type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    subject_id VARCHAR(255),
    workflow_id UUID,
    workflow_version VARCHAR(100),
    language VARCHAR(20),
    locale VARCHAR(20),
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_verifications_status ON verifications(status);
CREATE INDEX IF NOT EXISTS idx_verifications_subject_id ON verifications(subject_id);
CREATE INDEX IF NOT EXISTS idx_verifications_workflow_id ON verifications(workflow_id);
CREATE INDEX IF NOT EXISTS idx_verifications_created_at ON verifications(created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_verifications_external_id ON verifications(external_id) WHERE external_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS verification_audit_events (
    id UUID PRIMARY KEY,
    verification_id UUID NOT NULL REFERENCES verifications(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    actor_type VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255),
    previous_state VARCHAR(50),
    new_state VARCHAR(50),
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_verification_id ON verification_audit_events(verification_id);
