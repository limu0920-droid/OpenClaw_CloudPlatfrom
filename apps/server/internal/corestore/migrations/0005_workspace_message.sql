SET search_path TO platform, public;

CREATE TABLE IF NOT EXISTS workspace_message (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES workspace_session(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'recorded',
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workspace_message_session
    ON workspace_message (session_id, created_at ASC);
