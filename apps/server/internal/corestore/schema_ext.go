package corestore

import (
	"context"
	"database/sql"
)

const extendedSchemaSQL = `
CREATE TABLE IF NOT EXISTS platform.auth_identity (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES platform.user_account(id),
    tenant_id BIGINT NOT NULL REFERENCES platform.tenant(id),
    provider VARCHAR(32) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    status_reason TEXT,
    subject VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    open_id VARCHAR(255),
    union_id VARCHAR(255),
    external_name VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_auth_identity_provider_subject
    ON platform.auth_identity (provider, subject);

CREATE TABLE IF NOT EXISTS platform.channel (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    platform VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    connect_method VARCHAR(32) NOT NULL,
    auth_url TEXT,
    webhook_url TEXT,
    qrcode_url TEXT,
    token_masked TEXT,
    callback_secret TEXT,
    health_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    stats_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    entry_points_json JSONB NOT NULL DEFAULT '[]'::JSONB,
    settings_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    last_error TEXT,
    notes TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform.channel_activity (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL REFERENCES platform.channel(id) ON DELETE CASCADE,
    activity_type VARCHAR(32) NOT NULL,
    title VARCHAR(255) NOT NULL,
    summary TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform.account_settings (
    tenant_id BIGINT PRIMARY KEY REFERENCES platform.tenant(id),
    primary_email VARCHAR(255),
    billing_email VARCHAR(255),
    alert_email VARCHAR(255),
    preferred_locale VARCHAR(32) NOT NULL DEFAULT 'zh-CN',
    secondary_locale VARCHAR(32) NOT NULL DEFAULT 'en-US',
    timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    marketing_opt_in BOOLEAN NOT NULL DEFAULT FALSE,
    notify_on_alert BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_payment BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_expiry BOOLEAN NOT NULL DEFAULT TRUE,
    notify_channel_email BOOLEAN NOT NULL DEFAULT TRUE,
    notify_channel_webhook BOOLEAN NOT NULL DEFAULT FALSE,
    notify_channel_in_app BOOLEAN NOT NULL DEFAULT TRUE,
    notification_webhook_url TEXT,
    portal_headline TEXT,
    portal_subtitle TEXT,
    workspace_callout TEXT,
    experiment_badge VARCHAR(64),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE platform.account_settings
    ADD COLUMN IF NOT EXISTS notify_channel_email BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notify_channel_webhook BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS notify_channel_in_app BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notification_webhook_url TEXT,
    ADD COLUMN IF NOT EXISTS portal_headline TEXT,
    ADD COLUMN IF NOT EXISTS portal_subtitle TEXT,
    ADD COLUMN IF NOT EXISTS workspace_callout TEXT,
    ADD COLUMN IF NOT EXISTS experiment_badge VARCHAR(64);

CREATE TABLE IF NOT EXISTS platform.wallet_balance (
    tenant_id BIGINT PRIMARY KEY REFERENCES platform.tenant(id),
    currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    available_amount INTEGER NOT NULL DEFAULT 0,
    frozen_amount INTEGER NOT NULL DEFAULT 0,
    credit_limit INTEGER NOT NULL DEFAULT 0,
    auto_recharge BOOLEAN NOT NULL DEFAULT FALSE,
    last_settlement_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform.billing_statement (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES platform.tenant(id),
    statement_no VARCHAR(64) NOT NULL UNIQUE,
    billing_month VARCHAR(16) NOT NULL,
    status VARCHAR(32) NOT NULL,
    currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    opening_balance INTEGER NOT NULL DEFAULT 0,
    charge_amount INTEGER NOT NULL DEFAULT 0,
    refund_amount INTEGER NOT NULL DEFAULT 0,
    closing_balance INTEGER NOT NULL DEFAULT 0,
    paid_amount INTEGER NOT NULL DEFAULT 0,
    due_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform.ticket_record (
    id BIGSERIAL PRIMARY KEY,
    ticket_no VARCHAR(64) NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL REFERENCES platform.tenant(id),
    instance_id BIGINT REFERENCES platform.service_instance(id),
    title VARCHAR(255) NOT NULL,
    category VARCHAR(64) NOT NULL,
    severity VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    reporter VARCHAR(255) NOT NULL,
    assignee VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform.oem_brand (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    logo_url TEXT,
    favicon_url TEXT,
    support_email VARCHAR(255),
    support_url TEXT,
    domains_json JSONB NOT NULL DEFAULT '[]'::JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform.oem_theme (
    brand_id BIGINT PRIMARY KEY REFERENCES platform.oem_brand(id) ON DELETE CASCADE,
    primary_color VARCHAR(32),
    secondary_color VARCHAR(32),
    accent_color VARCHAR(32),
    surface_mode VARCHAR(32),
    font_family VARCHAR(128),
    radius VARCHAR(64)
);

CREATE TABLE IF NOT EXISTS platform.oem_feature_flags (
    brand_id BIGINT PRIMARY KEY REFERENCES platform.oem_brand(id) ON DELETE CASCADE,
    portal_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    admin_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    channels_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    tickets_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    purchase_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    runtime_control_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    audit_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sso_enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS platform.tenant_brand_binding (
    tenant_id BIGINT PRIMARY KEY REFERENCES platform.tenant(id),
    brand_id BIGINT NOT NULL REFERENCES platform.oem_brand(id),
    binding_mode VARCHAR(32) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS platform.payment_callback_event (
    id BIGSERIAL PRIMARY KEY,
    channel VARCHAR(32) NOT NULL,
    event_type VARCHAR(128) NOT NULL,
    out_trade_no VARCHAR(128),
    out_refund_no VARCHAR(128),
    signature_status VARCHAR(32) NOT NULL,
    decrypt_status VARCHAR(32) NOT NULL,
    process_status VARCHAR(32) NOT NULL,
    request_serial VARCHAR(128),
    raw_body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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

ALTER TABLE platform.order_main
    ADD COLUMN IF NOT EXISTS instance_id BIGINT,
    ADD COLUMN IF NOT EXISTS plan_code VARCHAR(64);

ALTER TABLE platform.subscription
    ADD COLUMN IF NOT EXISTS product_code VARCHAR(64),
    ADD COLUMN IF NOT EXISTS plan_code VARCHAR(64);

ALTER TABLE platform.payment_transaction
    ADD COLUMN IF NOT EXISTS currency VARCHAR(16) DEFAULT 'CNY',
    ADD COLUMN IF NOT EXISTS pay_mode VARCHAR(32),
    ADD COLUMN IF NOT EXISTS pay_url TEXT,
    ADD COLUMN IF NOT EXISTS code_url TEXT,
    ADD COLUMN IF NOT EXISTS prepay_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS app_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS mch_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS raw_json JSONB NOT NULL DEFAULT '{}'::JSONB;

ALTER TABLE platform.refund_record
    ADD COLUMN IF NOT EXISTS notify_url TEXT;

ALTER TABLE platform.invoice_record
    ADD COLUMN IF NOT EXISTS email VARCHAR(255),
    ADD COLUMN IF NOT EXISTS invoice_no VARCHAR(128),
    ADD COLUMN IF NOT EXISTS pdf_url TEXT;
`

func ensureExtendedSchema(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, extendedSchemaSQL)
	return err
}
