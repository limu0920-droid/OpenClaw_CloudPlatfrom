SET search_path TO platform, public;

ALTER TABLE workspace_artifact
    ADD COLUMN IF NOT EXISTS external_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS message_id BIGINT REFERENCES workspace_message(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS origin VARCHAR(32) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS archive_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS content_type VARCHAR(255),
    ADD COLUMN IF NOT EXISTS size_bytes BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS storage_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS storage_key TEXT,
    ADD COLUMN IF NOT EXISTS filename VARCHAR(255),
    ADD COLUMN IF NOT EXISTS checksum_sha256 VARCHAR(128);

CREATE INDEX IF NOT EXISTS idx_workspace_artifact_tenant_created
    ON workspace_artifact (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workspace_artifact_instance_created
    ON workspace_artifact (instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workspace_artifact_message
    ON workspace_artifact (message_id);

CREATE TABLE IF NOT EXISTS workspace_artifact_access_log (
    id BIGSERIAL PRIMARY KEY,
    artifact_id BIGINT NOT NULL REFERENCES workspace_artifact(id) ON DELETE CASCADE,
    session_id BIGINT NOT NULL REFERENCES workspace_session(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    action VARCHAR(32) NOT NULL,
    scope VARCHAR(32) NOT NULL,
    actor VARCHAR(255) NOT NULL,
    remote_addr VARCHAR(255),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workspace_artifact_access_log_artifact
    ON workspace_artifact_access_log (artifact_id, created_at DESC);
