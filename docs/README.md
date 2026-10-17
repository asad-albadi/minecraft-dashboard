# Documentation

- [Implementation plan](superpowers/plans/2026-10-17-minecraft-dashboard.md) — planned architecture, safety constraints, and verification tasks.
- [Troubleshooting](../TROUBLESHOOTING.md) — deployment and integration diagnosis.
- [Contributor guide](../AGENTS.md) — operating rules for developers and agents.

## Live read-only updates

`GET /api/v1/events` is a Server-Sent Events endpoint. It emits `connection.state`, `server.status`, and `console.line` events. Console events are taken from the existing read-only log path, passed through the same sanitization and RCON-noise filtering as the log-tail endpoint, and deduplicated across normal appends. On a log rotation the fresh sanitized tail is emitted so the dashboard resumes observing the replacement log. The endpoint never accepts commands and does not expose RCON credentials or host access.

## Player portraits

The browser renders an online player's public skin head from their public Minecraft name and sends no RCON credential or local player-data request to that image service. A CSS pixel-face remains beneath the image as a fallback for offline, missing, or blocked portraits.

Keep detailed architecture, adapter contracts, backup/restore runbooks, and security decisions in this directory as those integrations are implemented.
