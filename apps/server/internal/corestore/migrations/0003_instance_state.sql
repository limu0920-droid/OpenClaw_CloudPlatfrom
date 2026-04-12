CREATE TABLE IF NOT EXISTS platform.instance_runtime_state (
    instance_id BIGINT PRIMARY KEY REFERENCES platform.service_instance(id) ON DELETE CASCADE,
    power_state VARCHAR(32) NOT NULL,
    cpu_usage_percent INTEGER NOT NULL DEFAULT 0,
    memory_usage_percent INTEGER NOT NULL DEFAULT 0,
    disk_usage_percent INTEGER NOT NULL DEFAULT 0,
    api_requests_24h INTEGER NOT NULL DEFAULT 0,
    api_tokens_24h INTEGER NOT NULL DEFAULT 0,
    last_seen_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform.instance_credential (
    instance_id BIGINT PRIMARY KEY REFERENCES platform.service_instance(id) ON DELETE CASCADE,
    admin_user VARCHAR(128) NOT NULL,
    password_masked VARCHAR(255) NOT NULL,
    last_rotated_at TIMESTAMPTZ,
    requires_reset BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
