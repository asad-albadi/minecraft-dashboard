# Minecraft Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an additive Minecraft administration dashboard that matches the supplied pixel-art reference while safely integrating with an existing server only when configured.

**Architecture:** A Vite React/TypeScript single-page application renders the pixel-themed responsive dashboard and calls a Go/Fiber API. The API keeps integrations capability-gated: it reads status/logs/world metadata, uses RCON only through configured credentials, and performs lifecycle/backup work only through constrained adapters. SQLite persists chat, audit, and backup job metadata.

**Tech Stack:** Bun, Vite, React, TypeScript, CSS, Go, Fiber, SQLite, Docker Compose.

**Spec:** `MINECRAFT_DASHBOARD_SPEC.md`

## Global Constraints

- Preserve all existing Minecraft world, player data, configuration, and persistent volumes.
- Never expose RCON publicly, host shell access, secrets, or an unrestricted Docker socket.
- Return `Unavailable`/`Unknown` for unsupported or unconfigured metrics; never fabricate game data.
- Use real HTML text and original themed CSS/assets, not cropped UI text from the reference image.
- Require authentication, authorization, CSRF protection for cookie sessions, validation, rate limits, audit records, and confirmation for destructive operations.
- Keep Docker/log/config/world metadata access read-only by default; lifecycle and backup controls require explicit capability configuration.

## Review Focus

- Malicious chat characters must be JSON encoded for `tellraw`, never concatenated into RCON commands.
- Missing Docker/RCON/world paths must disable controls and report capabilities without failing the dashboard.
- Rotated logs must continue emitting ordered, deduplicated chat and console events.
- A viewer role must be denied all mutations, including chat and command submission.
- Failed backup save coordination must always re-enable saving and retain the active world untouched.

## File Structure

- `frontend/`: Bun/Vite React application, routes, API client, reusable pixel UI, responsive CSS, and frontend tests.
- `backend/cmd/dashboard/main.go`: configuration, Fiber setup, dependency wiring, and graceful shutdown.
- `backend/internal/`: domain models, authentication, adapters, HTTP handlers, WebSocket hub, persistence, and tests.
- `compose.yaml`, `.env.example`: additive deployment configuration only.
- `README.md`: setup, safety boundaries, configuration, deployment, rollback, and troubleshooting.

### Task 1: Bootstrap an additive full-stack workspace and safe capability model

**Files:**
- Create: `frontend/package.json`, `frontend/vite.config.ts`, `frontend/src/main.tsx`, `frontend/src/App.tsx`
- Create: `backend/go.mod`, `backend/cmd/dashboard/main.go`, `backend/internal/config/config.go`
- Create: `backend/internal/capabilities/service.go`, `backend/internal/capabilities/service_test.go`
- Create: `.env.example`, `compose.yaml`, `.gitignore`

**Interfaces:**
- Produces `CapabilityService.Snapshot() CapabilitySnapshot` and `GET /api/v1/capabilities`.
- Consumes only environment configuration; no Minecraft path is changed or created.

- [ ] **Step 1: Write failing capability tests** for absent RCON/Docker/world configuration producing disabled capabilities with reasons.
- [ ] **Step 2: Run `go test ./internal/capabilities`** and verify the tests fail because the service does not exist.
- [ ] **Step 3: Implement configuration parsing and `CapabilityService.Snapshot() CapabilitySnapshot`** with explicit disabled states.
- [ ] **Step 4: Run `go test ./internal/capabilities`** and verify it passes.
- [ ] **Step 5: Scaffold the Bun/Vite React application and Fiber health/capabilities routes**, preserving all secrets in server environment variables.
- [ ] **Step 6: Run `bun run build` in `frontend` and `go test ./...` in `backend`** and verify both pass.

### Task 2: Build the reference-faithful visual foundation and overview layout

**Files:**
- Create: `frontend/src/styles/tokens.css`, `frontend/src/styles/dashboard.css`
- Create: `frontend/src/components/{StonePanel,StoneInset,PanelHeader,PixelButton,PixelProgress,PixelAvatar,StatusLamp}.tsx`
- Create: `frontend/src/pages/OverviewPage.tsx`, `frontend/src/components/{Masthead,Sidebar,Hero,KpiStrip,WorldInfo,WorldWeather,QuickActions,PlayersPanel,ConsolePanel,Footer}.tsx`
- Create: `frontend/src/components/*.test.tsx`

**Interfaces:**
- Consumes API state types from `frontend/src/lib/api.ts`.
- Produces responsive overview composition and reusable pixel-themed primitives.

- [ ] **Step 1: Write failing component tests** that assert live/unknown state text, selected navigation, accessible buttons, and Console/Chat tab persistence.
- [ ] **Step 2: Run the focused frontend tests** and verify they fail before components exist.
- [ ] **Step 3: Implement semantic, keyboard-accessible pixel components and the 1536px three-column overview**, using CSS-created textures/frames and separately sourced original scenic art placeholders.
- [ ] **Step 4: Implement responsive breakpoints** for drawer navigation, 2×2 mobile KPIs, and stacked rails without clipped controls.
- [ ] **Step 5: Run focused tests and `bun run build`** and verify passing output.

### Task 3: Implement read-only server status, metadata, logs, and live updates

**Files:**
- Create: `backend/internal/status/{service,service_test}.go`, `backend/internal/world/{service,service_test}.go`
- Create: `backend/internal/logs/{tailer,parser,tailer_test,parser_test}.go`
- Create: `backend/internal/events/{hub,hub_test}.go`, `backend/internal/http/read_handlers.go`
- Modify: `frontend/src/lib/api.ts`, `frontend/src/pages/OverviewPage.tsx`

**Interfaces:**
- Produces `GET /api/v1/{status,metrics,players,world,logs}`, `WS /api/v1/events` and typed `console.line`, `server.status`, `server.metrics` events.
- Consumes configured read-only paths and optional status adapters.

- [ ] **Step 1: Write failing tests** for unavailable metrics, Minecraft chat parsing, initial log tail, log rotation, event ordering, and reconnect cursor replay.
- [ ] **Step 2: Run the backend package tests** and verify the expected missing-implementation failures.
- [ ] **Step 3: Implement read-only adapters and WebSocket hub** with timestamps, stable event IDs, bounded reconnect backoff, and honest unavailable responses.
- [ ] **Step 4: Bind live state into frontend panels** with themed loading, stale, error, and disconnected views.
- [ ] **Step 5: Run `go test ./...`, frontend tests, and `bun run build`** and verify passing output.

### Task 4: Add authenticated RCON console and two-way Minecraft chat

**Files:**
- Create: `backend/internal/auth/{service,middleware,service_test}.go`
- Create: `backend/internal/rcon/{client,client_test}.go`
- Create: `backend/internal/chat/{service,repository,service_test}.go`
- Create: `backend/internal/audit/service.go`, `backend/internal/http/mutation_handlers.go`
- Modify: `frontend/src/components/ConsolePanel.tsx`, `frontend/src/lib/api.ts`

**Interfaces:**
- Produces authenticated admin-only `POST /api/v1/chat/messages` and `POST /api/v1/console/commands`.
- `ChatService.Send(ctx, actor, message) (Message, error)` validates length/rate limits and encodes a `tellraw` JSON component.

- [ ] **Step 1: Write failing tests** for viewer denial, control-character rejection, quote/backslash encoding, RCON timeout, audit recording, and outgoing failure states.
- [ ] **Step 2: Run focused tests** and verify they fail for missing services.
- [ ] **Step 3: Implement session auth, role checks, CSRF checks, allowlisted RCON commands, SQLite chat/audit persistence, and safe `tellraw` injection.**
- [ ] **Step 4: Implement console/chat controls** with explicit delivery feedback, retained scroll behavior, and live event updates.
- [ ] **Step 5: Run `go test ./...`, frontend tests, and `bun run build`** and verify passing output.

### Task 5: Add capability-gated lifecycle and backup management

**Files:**
- Create: `backend/internal/lifecycle/{service,service_test}.go`
- Create: `backend/internal/backups/{service,service_test}.go`
- Create: `backend/internal/jobs/{service,service_test}.go`
- Create: `frontend/src/components/{ConfirmDialog,QuickActions}.tsx`, `frontend/src/pages/BackupsPage.tsx`

**Interfaces:**
- Produces admin-only lifecycle/backup APIs and job-progress events.
- `BackupService.Create(ctx, actor) (Job, error)` coordinates save-off/save-all flush/save-on in a `defer` cleanup path.

- [ ] **Step 1: Write failing tests** for confirmation requirement, disabled adapters, failing archive cleanup, save-on finally behavior, and path traversal rejection.
- [ ] **Step 2: Run focused backend tests** and verify they fail for absent lifecycle/backup services.
- [ ] **Step 3: Implement constrained adapter interfaces, progress jobs, audit events, checksums, retention metadata, and guarded restore workflow.**
- [ ] **Step 4: Implement confirmation dialogs and backup listing/status UI** with unavailable states when no explicit adapter is enabled.
- [ ] **Step 5: Run `go test ./...`, frontend tests, and `bun run build`** and verify passing output.

### Task 6: Build themed detail pages and operational documentation

**Files:**
- Create: `frontend/src/pages/{PlayersPage,WorldPage,ConsolePage,SettingsPage,BackupsPage}.tsx`
- Modify: `frontend/src/App.tsx`, `frontend/src/components/Sidebar.tsx`
- Create: `README.md`
- Modify: `compose.yaml`, `.env.example`

**Interfaces:**
- Consumes the REST/WebSocket APIs from Tasks 1–5.
- Produces routes `/players`, `/world`, `/console`, `/backups`, and `/settings` using the same components and capability state.

- [ ] **Step 1: Write failing route tests** for navigation, unavailable integration states, seed reveal permission, and destructive-action confirmations.
- [ ] **Step 2: Run focused frontend tests** and verify expected failures.
- [ ] **Step 3: Implement the pages and route integrations**, keeping advanced player/world features explicitly capability-gated.
- [ ] **Step 4: Document configuration, local development, deployment, rollback, backup/restore, security boundaries, and troubleshooting.**
- [ ] **Step 5: Run frontend tests/build and `go test ./...`** and verify passing output.

### Task 7: End-to-end verification and visual comparison

**Files:**
- Create: `frontend/e2e/dashboard.spec.ts`
- Create: `backend/internal/integration/dashboard_test.go`
- Modify: `README.md`

- [ ] **Step 1: Write failing E2E tests** for dashboard navigation, console/chat tabs, chat send feedback, offline state, and confirmation cancellation.
- [ ] **Step 2: Run the E2E suite** and verify it fails before the scenarios are implemented/configured.
- [ ] **Step 3: Add test fixtures and integration seams** without embedding fake production metrics.
- [ ] **Step 4: Run complete Go, frontend, E2E, and production build commands** and verify passing output.
- [ ] **Step 5: Capture a 1536×960 overview screenshot and compare it side-by-side with `minecraft-dashboard-reference.png`; correct material layout drift.**

## Self-Review

Coverage maps visual foundation to Task 2, read-only/live behavior to Task 3, chat/console/authentication to Task 4, lifecycle/backups to Task 5, detail pages/docs/deployment to Task 6, and final E2E/visual comparison to Task 7. No task assumes host paths, server credentials, or Docker control; each remains explicitly configured and disabled otherwise. The five highest-risk unspecified inputs are pinned by tests in Tasks 3–5.
