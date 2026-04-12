SET search_path TO platform, public;

CREATE TABLE IF NOT EXISTS workspace_session (
    id BIGSERIAL PRIMARY KEY,
    session_no VARCHAR(64) NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    workspace_url TEXT,
    last_opened_at TIMESTAMPTZ,
    last_artifact_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workspace_session_instance
    ON workspace_session (instance_id, created_at DESC);

CREATE TABLE IF NOT EXISTS workspace_artifact (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES workspace_session(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    kind VARCHAR(32) NOT NULL,
    source_url TEXT NOT NULL,
    preview_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workspace_artifact_session
    ON workspace_artifact (session_id, created_at DESC);
