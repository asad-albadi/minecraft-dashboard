# Family Generational Minecraft Server Dashboard

This is an additive dashboard stack for an existing Minecraft server. It does **not** create, mount over, reset, or delete the Minecraft server's world or player volumes.

## Development

The frontend uses Bun, Vite, React, and TypeScript. The backend uses Go and Fiber.

```sh
cd frontend && bun install && bun run dev
cd backend && go test ./... && go run ./cmd/dashboard
```

Copy `.env.example` to a private `.env` file before deployment. Leave integration variables blank for a safe visual/read-only shell. RCON, Docker lifecycle, backups, and world/log paths are deliberately opt-in and disabled by default.

## Safety boundaries

- Keep RCON on a private network; never publish its port.
- Do not mount `/var/run/docker.sock` into the public-facing dashboard.
- Mount Minecraft logs/config/world paths read-only. Use a dedicated backup destination, never the live world directory.
- Use a narrow local helper for lifecycle control, limited to the configured Minecraft container.
- Store real secrets only in deployment secret storage or an ignored `.env`; never commit them.
- Test restores in an isolated environment. A production restore must stop the server gracefully and create a pre-restore safety copy first.

## Deployment and rollback

`docker compose up -d --build` starts dashboard containers separately from Minecraft. The sample Compose file exposes services only on localhost. To roll back, stop and remove the dashboard stack; it has no ownership of Minecraft volumes. Preserve the `dashboard-data` volume only if you want dashboard audit/chat metadata retained.

## Troubleshooting

`UNKNOWN` status, disabled action buttons, and `Unavailable` metrics mean the related private adapter has not been configured or cannot reach its data source. They are intentionally not represented as invented server values.

For common configuration and recovery issues, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md). Project conventions are in [AGENTS.md](AGENTS.md), and longer-lived technical documentation lives in [docs/](docs/README.md).
