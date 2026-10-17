# Troubleshooting

## The dashboard says `UNKNOWN` or metrics are unavailable

This is expected until a verified status/metrics adapter is configured. The dashboard intentionally does not invent online state, player counts, TPS, memory, or world metadata. Check the dashboard backend logs and the configured private Minecraft host/path values.

## Quick actions or Chat/Console are disabled

They require explicit private configuration. Set the applicable values in an ignored `.env` file based on `.env.example`; never commit it. RCON must be reachable only from the backend on a private network. Lifecycle management additionally requires a narrowly scoped helper for the one configured Minecraft container — do not mount an unrestricted Docker socket.

## Frontend cannot reach the API

In development, run the Go backend on port 8080 and Vite proxies `/api` requests. In Compose, the Nginx frontend proxies `/api` to `dashboard-backend`. Confirm both services are running with `docker compose ps` and use the frontend’s localhost port 8081.

## Backend cannot reach the Minecraft `mc` service

The dashboard Compose stack joins the existing external `minecraft_default` Docker network so the backend can use the Minecraft service alias `mc` and its private RCON port. Confirm the Minecraft stack is running and that the network exists with `docker network inspect minecraft_default`. Do not publish port `25575` to the host or internet.

## World fields are unavailable

Set both `MINECRAFT_WORLD_PATH` and `MINECRAFT_PROPERTIES_PATH` in the ignored `.env` file to the existing host paths. Compose mounts them read-only inside the dashboard container; it never changes the world files. The seed is intentionally masked by default.

Minecraft time and weather additionally require the configured private RCON account to accept the vanilla read-only `time query daytime` and `weather query` commands. If a server/plugin returns a nonstandard response, the dashboard shows `Unavailable` rather than guessing; verify the commands directly from a trusted Minecraft administrator session before adding a compatibility adapter.

## Docker Compose reports missing or empty configuration

Copy `.env.example` to `.env`, fill only the integrations you have explicitly approved, and keep it outside version control. Blank optional settings are valid; they disable the matching capability safely.

## A container build fails because a local toolchain is missing

The Dockerfiles supply the Bun and Go build environments. Run `docker build ./frontend` or `docker build ./backend` to verify builds without installing either tool locally.

## Recovery and rollback

Stopping the dashboard stack is safe: `docker compose down`. It does not own or alter Minecraft volumes. Before enabling real backup/restore support, test restoration against a separate copy of the world and retain a verified pre-restore safety copy.
