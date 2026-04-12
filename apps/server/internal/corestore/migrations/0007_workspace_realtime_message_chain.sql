SET search_path TO platform, public;

ALTER TABLE workspace_session
    ADD COLUMN IF NOT EXISTS protocol_version VARCHAR(64) NOT NULL DEFAULT 'openclaw-lobster-bridge/v2',
    ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ;

ALTER TABLE workspace_message
    ADD COLUMN IF NOT EXISTS parent_message_id BIGINT REFERENCES workspace_message(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS external_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS origin VARCHAR(32) NOT NULL DEFAULT 'platform',
    ADD COLUMN IF NOT EXISTS trace_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS error_code VARCHAR(64),
    ADD COLUMN IF NOT EXISTS error_message TEXT,
    ADD COLUMN IF NOT EXISTS delivery_attempt INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_workspace_message_parent
    ON workspace_message (parent_message_id, created_at ASC);

CREATE UNIQUE INDEX IF NOT EXISTS uk_workspace_message_session_external_id
    ON workspace_message (session_id, external_id)
    WHERE external_id IS NOT NULL AND external_id <> '';

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

CREATE UNIQUE INDEX IF NOT EXISTS uk_workspace_artifact_session_external_id
    ON workspace_artifact (session_id, external_id)
    WHERE external_id IS NOT NULL AND external_id <> '';

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

CREATE TABLE IF NOT EXISTS workspace_message_event (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES workspace_session(id) ON DELETE CASCADE,
    message_id BIGINT REFERENCES workspace_message(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    external_id VARCHAR(128),
    origin VARCHAR(32),
    trace_id VARCHAR(128),
    payload_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workspace_message_event_session
    ON workspace_message_event (session_id, id ASC);

CREATE INDEX IF NOT EXISTS idx_workspace_message_event_message
    ON workspace_message_event (message_id, id ASC);
