SET search_path TO platform, public;

CREATE TABLE IF NOT EXISTS diagnostic_session (
    id BIGSERIAL PRIMARY KEY,
    session_no VARCHAR(64) NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    cluster_id VARCHAR(64),
    namespace VARCHAR(128),
    workload_id VARCHAR(128),
    workload_name VARCHAR(255),
    pod_name VARCHAR(255) NOT NULL,
    container_name VARCHAR(255),
    access_mode VARCHAR(32) NOT NULL DEFAULT 'readonly',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    approval_ticket VARCHAR(128),
    approved_by VARCHAR(255),
    operator VARCHAR(255) NOT NULL,
    operator_user_id BIGINT,
    reason TEXT,
    close_reason TEXT,
    expires_at TIMESTAMPTZ,
    last_command_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_diagnostic_session_instance
    ON diagnostic_session (instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_diagnostic_session_status
    ON diagnostic_session (status, updated_at DESC);

CREATE TABLE IF NOT EXISTS diagnostic_command_record (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES diagnostic_session(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    command_key VARCHAR(64),
    command_text TEXT NOT NULL,
    status VARCHAR(32) NOT NULL,
    exit_code INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    output TEXT,
    error_output TEXT,
    output_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_diagnostic_command_session
    ON diagnostic_command_record (session_id, executed_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_diagnostic_command_instance
    ON diagnostic_command_record (instance_id, executed_at DESC);
