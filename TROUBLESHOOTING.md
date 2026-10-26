# Troubleshooting

## Accounts are not configured (HTTP 503)

Shared access-secret login was replaced with named admin/viewer accounts. SESSION_SECRET is now a signing key of at least 32 bytes, not a password. Generate the private account file and configure its read-only mount as described in [setup](docs/setup.md). The file must be regular JSON, mode 0600/0400, readable by the backend UID, with unique valid account records and supported Argon2id hashes. Never loosen Minecraft source permissions to solve dashboard access. Restart only dashboard-backend after editing accounts.

## Login rejected, missing cookie, or HTTP 403/429

Use the frontend origin, not the backend port. Cookie Secure must be false only for localhost HTTP. For HTTPS use true and the exact DASHBOARD_PUBLIC_ORIGIN, without trailing slash. Login/logout/actions require Origin matching that value and JSON Content-Type. Five login attempts per source IP per minute are allowed; through the supplied proxy the limit is shared. Wait one minute after a 429. Session expiry, logout or backend restart requires signing in again.

## Unknown state or unavailable metrics

A failed RCON connection is not proof that Minecraft is offline. Java status ping and RCON are checked independently; when neither confirms reachability, state is Unknown. Set MINECRAFT_HOST (no scheme or port), MINECRAFT_PORT and the correct existing MINECRAFT_NETWORK. Do not expose private RCON or change the Minecraft container. CPU/memory/uptime/TPS/MSPT are unavailable without a reliable restricted metrics adapter.

## Chat / console / server controls are disabled

Viewer accounts cannot send actions. Admin chat requires DASHBOARD_ENABLE_CHAT=true; the restricted read-only console requires DASHBOARD_ENABLE_CONSOLE=true. Both require private RCON and writable dedicated audit storage. No toggle enables lifecycle, player/world mutations, backup writes or restore. An audit error before sending prevents delivery; a post-send failure reports unknown delivery, so check before retrying.

## API unavailable or old UI still displayed

Vite proxies /api to localhost:8080. Compose Nginx proxies to dashboard-backend:8080. Run docker compose ps. Rebuild only dashboard services with docker compose up -d --build dashboard-backend dashboard-frontend, then refresh the browser. This never restarts Minecraft.

## Console reconnects or history is missing

Nginx and any upstream proxy must not buffer text/event-stream. The shared SSE endpoint emits ordered console.snapshot events, so polling cannot duplicate a streamed line. The console retains 500 lines; reads are bounded and exceptionally long lines are dropped. Empty/unreadable log sources are unavailable, not fabricated output. On startup the bounded tail is reconciled against the latest retained event; entries outside that window during downtime cannot be recovered, and identical dateless lines across rotations can be ambiguous.

Activity/chat history needs a writable, dedicated DASHBOARD_DATA_PATH (Compose uses /var/lib/minecraft-dashboard). It must not overlap any source path or contain symlinks. A corrupt history.json is not overwritten; preserve a copy, inspect it offline and restore a known-good dashboard snapshot. Invalid storage falls back to memory and is reported in Settings. Audit storage failures prevent enabled actions.

## Player stats or advancements are Unknown

Sources are read from standard world/stats and world/advancements, with legacy world/players paths supported. All source mounts are read-only. Unknown missing keys are not zero. UUIDs are visible to authenticated users; precise positions are not exposed. First/last seen and joins reflect only retained recognizable logs. Date-only observations remain server dates, not converted fictitious timestamps. Malformed/oversized files, symlinked descendants and unsafe UUIDs are rejected. Public skin images may fall back to local artwork when mc-heads.net is blocked.

## World time, weather or version is unavailable

Status ping supplies the version when reachable; retained startup logs are a fallback. Vanilla command/version differences are expected. The dashboard queries daytime separately from day count; unsupported responses remain unavailable. Weather is never inferred from a previous weather-setting command. Size/dimensions refresh at most once per five minutes; static properties are labeled configured, not live.

## Backup catalog unavailable

Set an explicit backup source mount on the dashboard backend and its container path in DASHBOARD_BACKUP_PATH. Listing is read-only, skips symlinks and rejects directories above 1000 entries. File presence is not integrity or consistency. See [backup runbook](docs/backups.md).

## Desktop controls or console appear clipped

Desktop page scrolling is intentionally disabled. Detail pages, lower cards and console scroll inside their panels; short desktops allow the center column to scroll internally. Tablet/mobile pages reflow. Hard refresh after rebuilding the frontend. Keyboard focus can reach all scroll areas and controls.

## Recovery

Stop only the dashboard stack with docker compose down. Preserve dashboard-owned history/audit data for rollback. Never remove Minecraft volumes, mount over world data or test restoration against the live world.
