# Troubleshooting

## The dashboard says `UNKNOWN` or metrics are unavailable

This is expected until a verified status/metrics adapter is configured. The dashboard intentionally does not invent online state, player counts, TPS, memory, or world metadata. Check the dashboard backend logs and the configured private Minecraft host/path values.

## Quick actions or Chat/Console are disabled

They require explicit private configuration. Set the applicable values in an ignored `.env` file based on `.env.example`; never commit it. RCON must be reachable only from the backend on a private network. Lifecycle management additionally requires a narrowly scoped helper for the one configured Minecraft container — do not mount an unrestricted Docker socket.

## Frontend cannot reach the API

In development, run the Go backend on port 8080 and Vite proxies `/api` requests. In Compose, the Nginx frontend proxies `/api` to `dashboard-backend`. Confirm both services are running with `docker compose ps` and use the frontend’s localhost port 8081.

## The console says `RECONNECTING` or does not update live

The dashboard uses a same-origin Server-Sent Events (SSE) connection at `/api/v1/events` for read-only status and sanitized log updates. Confirm the backend is reachable through the frontend proxy and that no reverse proxy is buffering `text/event-stream` responses. The browser reconnects automatically; a persistent `RECONNECTING` label means the event endpoint is unavailable or the network path is interrupting the stream. Initial console history still loads from the read-only log-tail endpoint.

## A player portrait shows the fallback face

The player panel requests a public Minecraft head render by the public in-game name and uses the dashboard's built-in pixel face if that service is offline, blocked, or has no skin for the name. This affects only the browser image; it does not change, read, or expose Minecraft player-data files.

## The Players view is empty

The detailed profile cards are optional and read the live vanilla `usercache.json`, `ops.json`, and existing world `players/stats/*.json` sources. Set `MINECRAFT_USERCACHE_PATH` and `MINECRAFT_OPS_PATH` in the ignored `.env` file and recreate the dashboard backend. The dashboard intentionally omits UUIDs and exact locations, and every source mount is read only.

## The desktop dashboard does not scroll as a page

This is intentional. Desktop dashboards use a contained game-canvas layout; scroll the server-console log within its panel. Tablet and mobile layouts use normal page scrolling so controls remain reachable on narrow screens.

## The version badge says `Checking server…`

The badge reads the latest vanilla startup version from the configured read-only Minecraft log. It remains in its neutral loading state until a matching `Starting minecraft server version …` line is available; restart the dashboard after correcting a missing or unreadable log mount.

## Backend cannot reach the Minecraft `mc` service

The dashboard Compose stack joins the existing external `minecraft_default` Docker network so the backend can use the Minecraft service alias `mc` and its private RCON port. Confirm the Minecraft stack is running and that the network exists with `docker network inspect minecraft_default`. Do not publish port `25575` to the host or internet.

## World fields are unavailable

Set both `MINECRAFT_WORLD_PATH` and `MINECRAFT_PROPERTIES_PATH` in the ignored `.env` file to the existing host paths. Compose mounts them read-only inside the dashboard container; it never changes the world files. The seed is intentionally masked by default.

Minecraft time requires the configured private RCON account to accept the server's vanilla read-only clock command. This deployment uses `time query day`; older releases may use `time query daytime`. Its current vanilla command set does not expose a read-only weather query, so weather is hidden rather than guessed. Verify a server's command response from a trusted Minecraft administrator session before adding a compatibility adapter.

## Docker Compose reports missing or empty configuration

Copy `.env.example` to `.env`, fill only the integrations you have explicitly approved, and keep it outside version control. Blank optional settings are valid; they disable the matching capability safely.

## A container build fails because a local toolchain is missing

The Dockerfiles supply the Bun and Go build environments. Run `docker build ./frontend` or `docker build ./backend` to verify builds without installing either tool locally.

## Recovery and rollback

Stopping the dashboard stack is safe: `docker compose down`. It does not own or alter Minecraft volumes. Before enabling real backup/restore support, test restoration against a separate copy of the world and retain a verified pre-restore safety copy.
