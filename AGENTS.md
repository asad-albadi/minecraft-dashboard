# Contributor Guide

## Product guardrails

- Treat `MINECRAFT_DASHBOARD_SPEC.md` as the behavior specification and `minecraft-dashboard-reference.png` as the visual authority.
- Never create, replace, delete, or mount over the existing Minecraft world or player-data volumes.
- Keep RCON private. Do not expose host shell access, secrets, or an unrestricted Docker socket through this project.
- Return `Unavailable` or `Unknown` for unsupported data. Never substitute reference-image values for real server data.
- Keep lifecycle, backup, and restore features capability-gated until their adapters have explicit, least-privilege configuration.

## Development

- Frontend: `cd frontend && bun install && bun run dev`; run `bun run build` and `bun test` before committing frontend work.
- Backend: `cd backend && go test ./...`; the backend Dockerfile is the supported fallback where Go is unavailable locally.
- Keep user-facing controls keyboard accessible and preserve the reference’s pixel-stone visual style; do not replace it with generic rounded SaaS components.

## Documentation and commits

- Update `CHANGELOG.md` for every user-visible change.
- Update `TROUBLESHOOTING.md` when a new failure mode, configuration requirement, or recovery procedure is discovered.
- Put durable technical material in `docs/`, and keep `README.md` as the concise entry point.
- Make focused commits after a tested logical unit. Do not commit `.env`, credentials, build output, or Minecraft data.
