# OpenClaw Platform API External Integration Guide

## 1. Scope

This guide is the production-facing companion to `openapi.yaml`. It covers the integrations that sit outside the core Portal/Admin UI codebase:

- browser or OEM clients using the Platform API
- authentication integrators using Keycloak or WeChat login
- Lobster workspace bridge implementers
- object storage and artifact preview operators
- payment callback operators using WeChat Pay

The live doc index is exposed at `GET /api/v1/docs/external`.

## 2. Authentication

### Keycloak / OIDC

Use this sequence for browser-based login:

1. Call `GET /api/v1/bootstrap` or `GET /api/v1/auth/config` to discover the active auth provider.
2. Call `GET /api/v1/auth/keycloak/url?redirect_uri=<url>&next=<path>` to obtain the login URL.
3. Let the browser complete the Keycloak login and return to `GET /api/v1/auth/callback`.
4. The callback writes the signed auth cookie and redirects to the `next` target or `KEYCLOAK_POST_LOGIN_REDIRECT_URL`.
5. Use `GET /api/v1/auth/session` or `GET /api/v1/auth/me` to confirm the active session.

Production notes:

- `KEYCLOAK_SESSION_SECRET` must be unique per environment.
- `KEYCLOAK_COOKIE_SECURE` should stay `true` in production.
- Portal and Admin workspace APIs rely on the signed auth cookie; there is no parallel bearer-token flow today.

### WeChat Login

WeChat login is optional and should only be enabled after all three values are present:

- `WECHAT_LOGIN_ENABLED=true`
- `WECHAT_LOGIN_APP_ID`
- `WECHAT_LOGIN_APP_SECRET`
- `WECHAT_LOGIN_REDIRECT_URL`

Flow:

1. Call `GET /api/v1/auth/wechat/url?redirect_uri=<url>&next=<path>`.
2. Redirect the browser to the returned QR login URL.
3. Handle the callback at `GET /api/v1/auth/wechat/callback`.
4. Reuse the same auth cookie and session inspection endpoints as the Keycloak flow.

## 3. Workspace APIs

Workspace APIs are exposed under both scopes:

- `/api/v1/portal/...`
- `/api/v1/admin/...`

The same path contract is used for both and is modeled in `openapi.yaml` through the `{scope}` path parameter.

Core sequence:

1. Create or list sessions with `GET/POST /api/v1/{scope}/instances/{instanceId}/workspace/sessions`.
2. Inspect the global session index with `GET /api/v1/{scope}/workspace/sessions`.
3. Append messages through `POST /api/v1/{scope}/workspace/sessions/{sessionId}/messages`.
4. Use `POST /api/v1/{scope}/workspace/sessions/{sessionId}/messages/stream` for platform-originated SSE responses.
5. Subscribe to `GET /api/v1/{scope}/workspace/sessions/{sessionId}/events` for backlog replay and live event delivery.
6. Attach artifacts with `POST /api/v1/{scope}/workspace/sessions/{sessionId}/artifacts`.

Realtime behavior:

- `events` uses `text/event-stream`
- backlog replay honors `after` and `Last-Event-ID`
- bridge-originated events are normalized into `message.delta`, `reasoning.delta`, `tool.*`, `artifact.created`, `message.completed`, `message.failed`

## 4. Lobster Workspace Bridge

OpenClaw freezes the workspace bridge protocol at `openclaw-lobster-bridge/v2`.

Production configuration surface:

- `WORKSPACE_BRIDGE_PATH`
- `WORKSPACE_BRIDGE_HEALTH_PATH`
- `WORKSPACE_BRIDGE_HISTORY_PATH`
- `WORKSPACE_BRIDGE_STREAM_PATH`
- `WORKSPACE_BRIDGE_REPORT_PATH`
- `WORKSPACE_BRIDGE_HEADER_NAME`
- `WORKSPACE_BRIDGE_TOKEN`
- `WORKSPACE_BRIDGE_PUBLIC_BASE_URL`
- `WORKSPACE_BRIDGE_TIMEOUT_SECS`
- `WORKSPACE_BRIDGE_RETRY_COUNT`
- `WORKSPACE_BRIDGE_RETRY_BACKOFF_MS`
- `WORKSPACE_BRIDGE_HISTORY_SYNC`

Operational expectations:

- message dispatch is initiated by the Platform API against the bridge `messages` endpoint
- history sync is initiated by the Platform API against the bridge `history` endpoint
- upstream SSE streams are consumed by the Platform API and rewritten into workspace events
- asynchronous bridge callbacks are accepted at `POST /api/v1/platform/workspace/report`

Headers:

- `X-OpenClaw-Protocol-Version: openclaw-lobster-bridge/v2`
- `X-OpenClaw-Request-Id: <request-id>`
- `${WORKSPACE_BRIDGE_HEADER_NAME}: <token>` when a token is configured

`WORKSPACE_BRIDGE_PUBLIC_BASE_URL` must be a publicly routable URL for the Platform API so the bridge can call back to `report`.

The detailed wire contract remains in `docs/龙虾工作台桥接协议.md`.

## 5. Object Storage And Artifact Preview

Artifact archive and preview in production require both storage credentials and public preview settings.

Required storage settings:

- `OBJECT_STORAGE_ENDPOINT`
- `OBJECT_STORAGE_BUCKET`
- `OBJECT_STORAGE_ACCESS_KEY`
- `OBJECT_STORAGE_SECRET_KEY`
- `OBJECT_STORAGE_REGION`
- `OBJECT_STORAGE_FORCE_PATH_STYLE`

Required preview settings in strict mode:

- `ARTIFACT_PREVIEW_PUBLIC_BASE_URL`
- `ARTIFACT_PREVIEW_ALLOW_PRIVATE_IP=false`
- `ARTIFACT_PREVIEW_ALLOWED_HOSTS` should list every trusted object storage or CDN hostname used by artifacts

Preview behavior:

- `web/html`: proxied and sandboxed
- `pdf`: proxied inline
- `pptx/docx/xlsx`: prefer `previewUrl` PDF/HTML/image/text renditions, otherwise fall back to controlled download
- archived artifacts are served from object storage before the original upstream URL

Production recommendation:

- use a dedicated preview domain or at least a stable public base URL pointing to the same `platform-api` service
- never use wildcard hosts in `ARTIFACT_PREVIEW_ALLOWED_HOSTS`

## 6. WeChat Pay Callback

The inbound callback endpoint is `POST /api/v1/callback/payment/wechatpay`.

Production configuration surface:

- `WECHATPAY_ENABLED`
- `WECHATPAY_BASE_URL`
- `WECHATPAY_MCH_ID`
- `WECHATPAY_APP_ID`
- `WECHATPAY_NOTIFY_URL`
- `WECHATPAY_REFUND_NOTIFY_URL`
- `WECHATPAY_SERIAL_NO`
- `WECHATPAY_PUBLIC_KEY_ID`
- `WECHATPAY_PUBLIC_KEY_PEM`
- `WECHATPAY_PRIVATE_KEY_PEM`
- `WECHATPAY_APIV3_KEY`
- `WECHATPAY_MODE`
- `WECHATPAY_SUB_MCH_ID`
- `WECHATPAY_SUB_APP_ID`

The callback handler accepts both:

- plain transaction or refund payloads used in local testing
- API v3 style notification envelopes with signature verification and AES-GCM decryption

If `WECHATPAY_ENABLED=true`, production should provide the full certificate and API v3 key set so signature verification and decryption can run.

## 7. Production Checklist

Before release:

1. Render `prod` with `set-config.ps1` and a digest-pinned image.
2. Run `preflight.ps1 -Overlay prod -Strict`.
3. Run `preflight.ps1 -Overlay prod -Bootstrap -Strict`.
4. Ensure `security/prod/externalsecret.yaml` resolves every required key prefix.
5. Publish the release bundle so operators and integrators can consume:
   - sanitized manifests
   - the production config matrix
   - this integration guide
   - `openapi.yaml`

After release:

1. Verify `/readyz`.
2. Verify `/api/v1/docs/external`.
3. Verify workspace SSE, artifact preview, and bridge callback reachability in the target environment.
