# Repository Guidelines

## Project Structure & Module Organization
This repository is split by app and by working notes. `apps/web` contains the Vue 3 + Vite frontend; keep route-level work in `src/features`, shared shells in `src/layouts`, reusable UI in `src/components`, and API/types helpers in `src/lib`. `apps/platform-api` is the Go backend prototype; `cmd/server` is the entrypoint, while `internal/httpapi`, `internal/runtimeadapter`, `internal/oidc`, `internal/wechatauth`, and `internal/wechatpay` hold domain code. `docs/research` and `docs/plans` store analysis and implementation notes. Treat `apps/web/dist` and `apps/web/node_modules` as generated output.

## Build, Test, and Development Commands
- `cd apps/web; npm install; npm run dev` starts the frontend dev server with `/api` proxied to `http://localhost:8080`.
- `cd apps/web; npm run build` runs `vue-tsc` and produces the production bundle.
- `cd apps/platform-api; go run ./cmd/server` launches the platform backend on port `8080`.
- `cd apps/platform-api; go test ./...` runs the backend test suite.

## Coding Style & Naming Conventions
Use UTF-8 for all reads and writes. In `apps/web`, prefer TypeScript SFCs with `<script setup lang="ts">`, 2-space indentation, single quotes, and PascalCase file names such as `OverviewPage.vue` or `PortalLayout.vue`. Keep feature code inside `src/features/<area>`. In Go, rely on `gofmt`; package names stay lowercase, exported identifiers use Go conventions, and new server code should stay inside `internal/...` unless it is a binary entrypoint.

## Testing Guidelines
Backend tests live beside the implementation as `*_test.go`, currently concentrated in `apps/platform-api/internal/httpapi`. Add focused regression tests for API behavior, payment flows, and runtime adapter changes. There is no frontend test runner configured yet, so web changes must at least pass `npm run build` and a quick manual route check against the platform API.

## Commit & Pull Request Guidelines
Match the existing Git history style: short type prefix plus colon, for example `add: 初始化 OpenClaw 平台控制台与后端骨架`. Keep each commit to one logical change. Pull requests should summarize scope, list commands run, link the relevant issue or `docs/plans/...` note, and include screenshots or request/response examples for UI or API changes.

## Configuration & Safety Notes
Keep secrets out of Git and update `.env.example` files when adding new configuration. Do not commit local logs, `dist/`, or `node_modules/` artifacts.


