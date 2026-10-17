# Troubleshooting

## The dashboard says `UNKNOWN` or metrics are unavailable

This is expected until a verified status/metrics adapter is configured. The dashboard intentionally does not invent online state, player counts, TPS, memory, or world metadata. Check the dashboard backend logs and the configured private Minecraft host/path values.

## Quick actions or Chat/Console are disabled

They require explicit private configuration. Set the applicable values in an ignored `.env` file based on `.env.example`; never commit it. RCON must be reachable only from the backend on a private network. Lifecycle management additionally requires a narrowly scoped helper for the one configured Minecraft container — do not mount an unrestricted Docker socket.

## Frontend cannot reach the API

In development, run the Go backend on port 8080 and Vite proxies `/api` requests. In Compose, the Nginx frontend proxies `/api` to `dashboard-backend`. Confirm both services are running with `docker compose ps` and use the frontend’s localhost port 8081.

## Docker Compose reports missing or empty configuration

Copy `.env.example` to `.env`, fill only the integrations you have explicitly approved, and keep it outside version control. Blank optional settings are valid; they disable the matching capability safely.

## A container build fails because a local toolchain is missing

The Dockerfiles supply the Bun and Go build environments. Run `docker build ./frontend` or `docker build ./backend` to verify builds without installing either tool locally.

## Recovery and rollback

Stopping the dashboard stack is safe: `docker compose down`. It does not own or alter Minecraft volumes. Before enabling real backup/restore support, test restoration against a separate copy of the world and retain a verified pre-restore safety copy.
