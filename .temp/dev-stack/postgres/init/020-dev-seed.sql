SET search_path TO platform, public;

INSERT INTO product (code, name, type, status, description)
VALUES ('openclaw', 'OpenClaw', 'saas', 'active', 'OpenClaw local dev product')
ON CONFLICT (code) DO NOTHING;

INSERT INTO service_plan (product_id, code, name, status, billing_mode, trial_supported, resource_spec, feature_spec)
SELECT p.id,
       'trial',
       'Trial',
       'active',
       'subscription',
       TRUE,
       '{"cpu":"1","memory":"2Gi","storage":"10Gi"}'::jsonb,
       '{"runtimeControl":true,"tickets":true}'::jsonb
FROM product p
WHERE p.code = 'openclaw'
ON CONFLICT (code) DO NOTHING;

INSERT INTO plan_price (plan_id, billing_cycle, currency, amount, status)
SELECT sp.id, 'monthly', 'CNY', 0, 'active'
FROM service_plan sp
WHERE sp.code = 'trial'
ON CONFLICT (plan_id, billing_cycle, currency) DO NOTHING;

INSERT INTO app_role (code, name, scope, status, permissions)
VALUES
  ('tenant_admin', 'Tenant Admin', 'tenant', 'active', '["instance:read","instance:write","billing:read"]'::jsonb),
  ('platform_admin', 'Platform Admin', 'platform', 'active', '["*"]'::jsonb)
ON CONFLICT (code) DO NOTHING;
