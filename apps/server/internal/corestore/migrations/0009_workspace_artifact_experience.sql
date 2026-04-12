SET search_path TO platform, public;

CREATE TABLE IF NOT EXISTS workspace_artifact_favorite (
    id BIGSERIAL PRIMARY KEY,
    artifact_id BIGINT NOT NULL REFERENCES workspace_artifact(id) ON DELETE CASCADE,
    session_id BIGINT NOT NULL REFERENCES workspace_session(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    user_id BIGINT,
    actor VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_workspace_artifact_favorite_actor
    ON workspace_artifact_favorite (artifact_id, COALESCE(user_id, 0), actor);

CREATE INDEX IF NOT EXISTS idx_workspace_artifact_favorite_tenant_created
    ON workspace_artifact_favorite (tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS workspace_artifact_share (
    id BIGSERIAL PRIMARY KEY,
    artifact_id BIGINT NOT NULL REFERENCES workspace_artifact(id) ON DELETE CASCADE,
    session_id BIGINT NOT NULL REFERENCES workspace_session(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    instance_id BIGINT NOT NULL REFERENCES service_instance(id) ON DELETE CASCADE,
    scope VARCHAR(32) NOT NULL,
    token VARCHAR(128) NOT NULL,
    note TEXT,
    created_by VARCHAR(255) NOT NULL,
    created_by_user_id BIGINT,
    use_count INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    last_opened_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_workspace_artifact_share_token
    ON workspace_artifact_share (token);

CREATE INDEX IF NOT EXISTS idx_workspace_artifact_share_artifact_created
    ON workspace_artifact_share (artifact_id, created_at DESC);
