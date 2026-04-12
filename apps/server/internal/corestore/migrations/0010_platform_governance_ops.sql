SET search_path TO platform, public;

ALTER TABLE approval_record
    ADD COLUMN IF NOT EXISTS instance_id BIGINT REFERENCES service_instance(id),
    ADD COLUMN IF NOT EXISTS executor_id BIGINT REFERENCES user_account(id),
    ADD COLUMN IF NOT EXISTS executed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS risk_level VARCHAR(16) NOT NULL DEFAULT 'high',
    ADD COLUMN IF NOT EXISTS reject_reason TEXT,
    ADD COLUMN IF NOT EXISTS metadata_json JSONB NOT NULL DEFAULT '{}'::JSONB;

UPDATE approval_record
SET risk_level = 'high'
WHERE risk_level IS NULL OR risk_level = '';

UPDATE approval_record
SET metadata_json = '{}'::JSONB
WHERE metadata_json IS NULL;

ALTER TABLE approval_record
    ALTER COLUMN risk_level SET DEFAULT 'high',
    ALTER COLUMN risk_level SET NOT NULL,
    ALTER COLUMN metadata_json SET DEFAULT '{}'::JSONB,
    ALTER COLUMN metadata_json SET NOT NULL;

CREATE TABLE IF NOT EXISTS approval_action (
    id BIGSERIAL PRIMARY KEY,
    approval_id BIGINT NOT NULL REFERENCES approval_record(id) ON DELETE CASCADE,
    actor_id BIGINT REFERENCES user_account(id),
    actor_name VARCHAR(255) NOT NULL,
    action VARCHAR(32) NOT NULL,
    comment TEXT,
    metadata_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_approval_action_approval_created
    ON approval_action (approval_id, created_at DESC);

CREATE TABLE IF NOT EXISTS diagnostic_session (
    id BIGSERIAL PRIMARY KEY,
    session_no VARCHAR(64) NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    cluster_id VARCHAR(64),
    namespace VARCHAR(128),
    workload_id VARCHAR(128),
    workload_name VARCHAR(128),
    pod_name VARCHAR(255) NOT NULL,
    container_name VARCHAR(128),
    access_mode VARCHAR(32) NOT NULL DEFAULT 'readonly',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    approval_ticket VARCHAR(64),
    approved_by VARCHAR(255),
    operator VARCHAR(255) NOT NULL,
    operator_user_id BIGINT REFERENCES user_account(id),
    reason TEXT,
    close_reason TEXT,
    expires_at TIMESTAMPTZ,
    last_command_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE diagnostic_session
    ADD COLUMN IF NOT EXISTS cluster_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS namespace VARCHAR(128),
    ADD COLUMN IF NOT EXISTS workload_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS workload_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS pod_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS container_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS access_mode VARCHAR(32) NOT NULL DEFAULT 'readonly',
    ADD COLUMN IF NOT EXISTS approval_ticket VARCHAR(128),
    ADD COLUMN IF NOT EXISTS approved_by VARCHAR(255),
    ADD COLUMN IF NOT EXISTS operator VARCHAR(255),
    ADD COLUMN IF NOT EXISTS operator_user_id BIGINT,
    ADD COLUMN IF NOT EXISTS reason TEXT,
    ADD COLUMN IF NOT EXISTS close_reason TEXT,
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_command_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS ended_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE diagnostic_session
    ALTER COLUMN workload_name TYPE VARCHAR(255),
    ALTER COLUMN container_name TYPE VARCHAR(255),
    ALTER COLUMN approval_ticket TYPE VARCHAR(128);

UPDATE diagnostic_session
SET started_at = COALESCE(started_at, created_at, NOW())
WHERE started_at IS NULL;

UPDATE diagnostic_session
SET updated_at = COALESCE(updated_at, created_at, started_at, NOW())
WHERE updated_at IS NULL;

ALTER TABLE diagnostic_session
    ALTER COLUMN started_at SET DEFAULT NOW(),
    ALTER COLUMN started_at SET NOT NULL,
    ALTER COLUMN updated_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_diagnostic_session_instance
    ON diagnostic_session (instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_diagnostic_session_status
    ON diagnostic_session (status, access_mode, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_diagnostic_session_status_access_mode
    ON diagnostic_session (status, access_mode, updated_at DESC);

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

ALTER TABLE diagnostic_command_record
    ADD COLUMN IF NOT EXISTS command_key VARCHAR(64),
    ADD COLUMN IF NOT EXISTS exit_code INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS duration_ms INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS output TEXT,
    ADD COLUMN IF NOT EXISTS error_output TEXT,
    ADD COLUMN IF NOT EXISTS output_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE diagnostic_command_record
SET output_truncated = FALSE
WHERE output_truncated IS NULL;

UPDATE diagnostic_command_record
SET executed_at = COALESCE(executed_at, NOW())
WHERE executed_at IS NULL;

ALTER TABLE diagnostic_command_record
    ALTER COLUMN output_truncated SET DEFAULT FALSE,
    ALTER COLUMN output_truncated SET NOT NULL,
    ALTER COLUMN executed_at SET DEFAULT NOW(),
    ALTER COLUMN executed_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_diagnostic_command_session
    ON diagnostic_command_record (session_id, executed_at ASC);

CREATE INDEX IF NOT EXISTS idx_diagnostic_command_instance
    ON diagnostic_command_record (instance_id, executed_at DESC);
