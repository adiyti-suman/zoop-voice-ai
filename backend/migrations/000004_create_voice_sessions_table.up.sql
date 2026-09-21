CREATE TABLE IF NOT EXISTS voice_sessions (
    id UUID PRIMARY KEY,
    verification_id UUID NOT NULL REFERENCES verifications(id) ON DELETE CASCADE,
    workflow_run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    language VARCHAR(50),
    sample_rate INT NOT NULL,
    channels INT NOT NULL,
    encoding VARCHAR(50) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_voice_sessions_verification_id ON voice_sessions(verification_id);
CREATE INDEX idx_voice_sessions_workflow_run_id ON voice_sessions(workflow_run_id);
