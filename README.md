# Minecraft Server Dashboard

An independently deployed, Minecraft-themed dashboard for any existing Java server. Monitor players, saved statistics, advancements, world metadata, logs and chat without modifying the server container or its data. Set `DASHBOARD_SERVER_NAME` for your own server identity; the default is **Minecraft Server**.

## Get started

1. Copy `.env.example` to a private `.env` and configure only existing, authorized read-only sources and the server's private network.
2. Generate named administrator/viewer accounts with `cd backend && go run ./cmd/setup-accounts -directory ../.dashboard`. Private credentials and Argon2id account hashes are written locally, never displayed or committed. See [setup and security](docs/setup.md) for Docker tooling and file ownership.
3. Set a random `SESSION_SECRET` (at least 32 bytes), the account source file and dashboard data directory. The signing secret is **not a login password**.
4. Run `docker compose up -d --build dashboard-backend dashboard-frontend`, then open the dashboard at `http://127.0.0.1:8081`.

The six pages provide monitoring, searchable player profiles, world diagnostics, searchable streaming logs and parsed chat, a read-only backup catalog, and browser-local preferences. Administrators and viewers have separate sessions; all Minecraft-changing features are disabled by default. Optional chat and a restricted read-only console require explicit configuration and durable audit storage.

## Development

The frontend uses Bun, Vite, React, and TypeScript. The backend uses Go and Fiber.

```sh
cd frontend && bun install && bun run dev
cd backend && go test ./... && go run ./cmd/dashboard
```

Run `bun run build` and `bun test` in `frontend`, and `go test ./...` plus `go vet ./...` in `backend`. Go is also available through the backend Docker build environment. Use `DASHBOARD_COOKIE_SECURE=false` only for local HTTP development.

## Safety boundaries

- Keep RCON on a private network; never publish its port.
- Do not mount `/var/run/docker.sock` into the public-facing dashboard.
- Mount Minecraft logs/config/world paths read-only. Use a dedicated backup destination, never the live world directory.
- Lifecycle, world changes and restore are unavailable without a separately authorized narrow adapter; this dashboard does not create one.
- Store real secrets only in deployment secret storage or an ignored `.env`; never commit them.
- Test backup verification/restore on an isolated copy; never overwrite a live world during dashboard testing. See the [backup runbook](docs/backups.md).

## Deployment and rollback

The sample Compose file exposes services only on localhost. Configure `MINECRAFT_NETWORK` to an existing network; no Minecraft service is declared or recreated. For HTTPS proxying, configure the exact public origin and secure cookies. To roll back, stop only this dashboard stack and restore its previous image/configuration. Preserve dashboard data for retained chat, activity and audit history. Account migration intentionally invalidates old shared-password sessions.

## Capability limits

Unknown values are not replaced with samples. Metrics, live inventory, complete session durations, map rendering, coordinated backup creation, backup downloads/deletion, retention scheduling and restoration are not implemented without their external integrations. Chat uses the existing SSE monitoring transport plus retained history polling, not a WebSocket connection. See [feature coverage](docs/feature-coverage.md) for implemented and deferred requirements.

## Troubleshooting

`UNKNOWN` status, disabled action buttons, and `Unavailable` metrics mean the related private adapter has not been configured or cannot reach its data source. They are intentionally not represented as invented server values.

For common configuration and recovery issues, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md). Project conventions are in [AGENTS.md](AGENTS.md), and longer-lived technical documentation lives in [docs/](docs/README.md).
