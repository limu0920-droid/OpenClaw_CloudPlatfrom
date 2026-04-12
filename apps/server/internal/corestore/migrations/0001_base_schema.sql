CREATE SCHEMA IF NOT EXISTS platform;

SET search_path TO platform, public;

CREATE TABLE IF NOT EXISTS product (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS service_plan (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES product(id),
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    billing_mode VARCHAR(32) NOT NULL DEFAULT 'subscription',
    trial_supported BOOLEAN NOT NULL DEFAULT FALSE,
    resource_spec JSONB NOT NULL DEFAULT '{}'::JSONB,
    feature_spec JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS plan_price (
    id BIGSERIAL PRIMARY KEY,
    plan_id BIGINT NOT NULL REFERENCES service_plan(id),
    billing_cycle VARCHAR(32) NOT NULL,
    currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    amount INTEGER NOT NULL CHECK (amount >= 0),
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (plan_id, billing_cycle, currency)
);

CREATE TABLE IF NOT EXISTS tenant (
    id BIGSERIAL PRIMARY KEY,
    tenant_code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    plan_id BIGINT REFERENCES service_plan(id),
    owner_user_id BIGINT,
    expired_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_account (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    login_name VARCHAR(128) NOT NULL,
    display_name VARCHAR(128),
    email VARCHAR(256),
    phone VARCHAR(32),
    password_hash TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ,
    profile JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, login_name)
);

CREATE TABLE IF NOT EXISTS app_role (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    scope VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    permissions JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_role_rel (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    user_id BIGINT NOT NULL REFERENCES user_account(id),
    role_id BIGINT NOT NULL REFERENCES app_role(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, role_id)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_tenant_owner_user'
    ) THEN
        ALTER TABLE tenant
            ADD CONSTRAINT fk_tenant_owner_user
            FOREIGN KEY (owner_user_id) REFERENCES user_account(id);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS cluster (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    region VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cluster_node (
    id BIGSERIAL PRIMARY KEY,
    cluster_id BIGINT NOT NULL REFERENCES cluster(id),
    hostname VARCHAR(128) NOT NULL,
    node_type VARCHAR(32) NOT NULL DEFAULT 'worker',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    capacity_cpu_milli INTEGER NOT NULL DEFAULT 0,
    capacity_memory_mb INTEGER NOT NULL DEFAULT 0,
    capacity_disk_gb INTEGER NOT NULL DEFAULT 0,
    labels JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cluster_id, hostname)
);

CREATE TABLE IF NOT EXISTS service_instance (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    cluster_id BIGINT NOT NULL REFERENCES cluster(id),
    plan_id BIGINT NOT NULL REFERENCES service_plan(id),
    instance_code VARCHAR(64) NOT NULL UNIQUE,
    display_name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'creating',
    version VARCHAR(64),
    runtime_type VARCHAR(32) NOT NULL DEFAULT 'docker',
    primary_node_id BIGINT REFERENCES cluster_node(id),
    resource_spec JSONB NOT NULL DEFAULT '{}'::JSONB,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    activated_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS runtime_container (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES service_instance(id),
    runtime_type VARCHAR(32) NOT NULL DEFAULT 'docker',
    container_ref VARCHAR(128) NOT NULL,
    container_name VARCHAR(128),
    node_id BIGINT REFERENCES cluster_node(id),
    status VARCHAR(32) NOT NULL DEFAULT 'running',
    image_ref VARCHAR(255),
    resource_limit_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (runtime_type, container_ref)
);

CREATE TABLE IF NOT EXISTS instance_member (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES service_instance(id),
    user_id BIGINT NOT NULL REFERENCES user_account(id),
    member_role VARCHAR(32) NOT NULL DEFAULT 'member',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (instance_id, user_id)
);

CREATE TABLE IF NOT EXISTS instance_access (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES service_instance(id),
    entry_type VARCHAR(32) NOT NULL,
    url TEXT NOT NULL,
    domain VARCHAR(255),
    access_mode VARCHAR(32) NOT NULL DEFAULT 'normal',
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS instance_config (
    instance_id BIGINT PRIMARY KEY REFERENCES service_instance(id),
    config_version INTEGER NOT NULL DEFAULT 1,
    config_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    config_hash VARCHAR(128),
    published_at TIMESTAMPTZ,
    updated_by BIGINT REFERENCES user_account(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS instance_config_history (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES service_instance(id),
    version INTEGER NOT NULL,
    change_type VARCHAR(32) NOT NULL DEFAULT 'publish',
    config_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    config_hash VARCHAR(128),
    changed_by BIGINT REFERENCES user_account(id),
    change_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (instance_id, version)
);

CREATE TABLE IF NOT EXISTS backup_record (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES service_instance(id),
    backup_no VARCHAR(64) NOT NULL UNIQUE,
    backup_type VARCHAR(32) NOT NULL DEFAULT 'manual',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    storage_uri TEXT,
    size_bytes BIGINT,
    expired_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_by BIGINT REFERENCES user_account(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS restore_record (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES service_instance(id),
    backup_id BIGINT NOT NULL REFERENCES backup_record(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    restore_scope VARCHAR(32) NOT NULL DEFAULT 'full',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_by BIGINT REFERENCES user_account(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_main (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    order_no VARCHAR(64) NOT NULL UNIQUE,
    source_platform VARCHAR(32) NOT NULL DEFAULT 'portal',
    external_order_no VARCHAR(128),
    order_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    total_amount INTEGER NOT NULL CHECK (total_amount >= 0),
    discount_amount INTEGER NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    payable_amount INTEGER NOT NULL DEFAULT 0 CHECK (payable_amount >= 0),
    paid_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_by BIGINT REFERENCES user_account(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_main_external_order_no
    ON order_main (external_order_no);

CREATE TABLE IF NOT EXISTS order_item (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES order_main(id),
    product_id BIGINT NOT NULL REFERENCES product(id),
    plan_id BIGINT NOT NULL REFERENCES service_plan(id),
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price INTEGER NOT NULL CHECK (unit_price >= 0),
    total_amount INTEGER NOT NULL CHECK (total_amount >= 0),
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscription (
    id BIGSERIAL PRIMARY KEY,
    subscription_no VARCHAR(64) NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    product_id BIGINT NOT NULL REFERENCES product(id),
    plan_id BIGINT NOT NULL REFERENCES service_plan(id),
    instance_id BIGINT REFERENCES service_instance(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    renew_mode VARCHAR(32) NOT NULL DEFAULT 'manual',
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_transaction (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES order_main(id),
    channel VARCHAR(32) NOT NULL,
    trade_no VARCHAR(128) NOT NULL UNIQUE,
    channel_order_no VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'created',
    amount INTEGER NOT NULL CHECK (amount >= 0),
    request_payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    callback_payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_payment_transaction_channel_order_no
    ON payment_transaction (channel, channel_order_no)
    WHERE channel_order_no IS NOT NULL;

CREATE TABLE IF NOT EXISTS refund_record (
    id BIGSERIAL PRIMARY KEY,
    refund_no VARCHAR(64) NOT NULL UNIQUE,
    order_id BIGINT NOT NULL REFERENCES order_main(id),
    payment_id BIGINT NOT NULL REFERENCES payment_transaction(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    amount INTEGER NOT NULL CHECK (amount >= 0),
    reason TEXT,
    requested_by BIGINT REFERENCES user_account(id),
    channel_refund_no VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoice_record (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenant(id),
    order_id BIGINT NOT NULL REFERENCES order_main(id),
    invoice_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    amount INTEGER NOT NULL CHECK (amount >= 0),
    title VARCHAR(255),
    tax_no VARCHAR(64),
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS operation_job (
    id BIGSERIAL PRIMARY KEY,
    job_no VARCHAR(64) NOT NULL UNIQUE,
    job_type VARCHAR(64) NOT NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT,
    tenant_id BIGINT REFERENCES tenant(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    priority SMALLINT NOT NULL DEFAULT 5,
    payload_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    result_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    error_message TEXT,
    scheduled_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_by BIGINT REFERENCES user_account(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS operation_log (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenant(id),
    operator_id BIGINT REFERENCES user_account(id),
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT,
    result VARCHAR(32) NOT NULL,
    request_id VARCHAR(64),
    content_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS terminal_session (
    id BIGSERIAL PRIMARY KEY,
    session_no VARCHAR(64) NOT NULL UNIQUE,
    instance_id BIGINT NOT NULL REFERENCES service_instance(id),
    container_ref VARCHAR(128) NOT NULL,
    operator_id BIGINT NOT NULL REFERENCES user_account(id),
    approved_by BIGINT REFERENCES user_account(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    approval_no VARCHAR(64),
    command_policy VARCHAR(32) NOT NULL DEFAULT 'readonly',
    record_uri TEXT,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS approval_record (
    id BIGSERIAL PRIMARY KEY,
    approval_no VARCHAR(64) NOT NULL UNIQUE,
    tenant_id BIGINT REFERENCES tenant(id),
    approval_type VARCHAR(64) NOT NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT,
    applicant_id BIGINT NOT NULL REFERENCES user_account(id),
    approver_id BIGINT REFERENCES user_account(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    reason TEXT,
    approval_comment TEXT,
    approved_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_record (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenant(id),
    user_id BIGINT REFERENCES user_account(id),
    notify_type VARCHAR(32) NOT NULL,
    channel VARCHAR(32) NOT NULL,
    template_code VARCHAR(64),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    payload_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    result_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alert_record (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenant(id),
    instance_id BIGINT REFERENCES service_instance(id),
    metric_key VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    summary TEXT NOT NULL,
    detail JSONB NOT NULL DEFAULT '{}'::JSONB,
    triggered_at TIMESTAMPTZ NOT NULL,
    acknowledged_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_event (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenant(id),
    operator_id BIGINT REFERENCES user_account(id),
    event_type VARCHAR(64) NOT NULL,
    risk_level VARCHAR(16) NOT NULL DEFAULT 'medium',
    target_type VARCHAR(32),
    target_id BIGINT,
    request_id VARCHAR(64),
    source_ip VARCHAR(64),
    content_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_status
    ON tenant (status, expired_at);

CREATE INDEX IF NOT EXISTS idx_user_account_tenant_status
    ON user_account (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_cluster_node_cluster_status
    ON cluster_node (cluster_id, status);

CREATE INDEX IF NOT EXISTS idx_service_instance_tenant_status
    ON service_instance (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_service_instance_cluster_status
    ON service_instance (cluster_id, status);

CREATE INDEX IF NOT EXISTS idx_instance_member_user
    ON instance_member (user_id, status);

CREATE INDEX IF NOT EXISTS idx_runtime_container_instance
    ON runtime_container (instance_id, status);

CREATE INDEX IF NOT EXISTS idx_instance_access_instance
    ON instance_access (instance_id, is_primary);

CREATE INDEX IF NOT EXISTS idx_config_history_instance
    ON instance_config_history (instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_backup_record_instance
    ON backup_record (instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_restore_record_instance
    ON restore_record (instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscription_tenant_product
    ON subscription (tenant_id, product_id, status);

CREATE INDEX IF NOT EXISTS idx_operation_job_status
    ON operation_job (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_operation_job_target
    ON operation_job (target_type, target_id);

CREATE INDEX IF NOT EXISTS idx_operation_log_tenant_created
    ON operation_log (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_terminal_session_instance
    ON terminal_session (instance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_approval_record_status
    ON approval_record (status, approval_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notification_record_status
    ON notification_record (status, notify_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_alert_record_status
    ON alert_record (status, severity, triggered_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_event_tenant_created
    ON audit_event (tenant_id, created_at DESC);
