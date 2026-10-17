# Family Generational Minecraft Server — Dashboard Implementation Specification

**Status:** Implementation-ready product and visual specification  
**Primary visual reference:** `minecraft-dashboard-reference.png` (must be provided alongside this document)  
**Priority:** Reproduce the reference image as faithfully as practical while delivering a real, secure Minecraft server administration application.  
**Scope:** Self-hosted single-server dashboard initially, extensible later.

> **Source-of-truth rule:** The supplied image is authoritative for appearance, geometry, ordering, visual density, textures, and typography. This document is authoritative for interaction and backend behavior. Where the image is illustrative or technically ambiguous, use the implementation rules below. Do not treat any illustrated server statistics, player names, version numbers, world seed, or chat lines as actual data.

## 1. Product purpose

A web-based, Minecraft-themed control center for the existing **Family Generational Minecraft Server**. An authenticated administrator can monitor the server, view players and live logs, communicate bidirectionally with players in-game, issue RCON commands, manage backups and server lifecycle, and inspect world information. This is a real application, **not a static recreation** of the screenshot.

The intended existing deployment is an `itzg/minecraft-server` Docker container on a Debian host. Preserve the existing world, player data, and configuration. Do not recreate, replace, or delete its persistent volumes as part of dashboard installation.

## 2. Exact visual reference and composition

### 2.1 Reference canvas

- Image dimensions: **1536 × 960 pixels**; widescreen desktop reference, ratio **1.6:1**.
- The interface occupies the full canvas edge-to-edge with no conventional white page margin.
- Overall appearance: Minecraft survival-world atmosphere, richly detailed pixel-art textures, deep charcoal stone, blackened iron frames, wood and grass-block motifs, lanterns and warm amber highlights.
- The look is **skeuomorphic Minecraft UI**, not modern glassmorphism or a generic Tailwind admin template. Use sharp/pixelated shapes, stepped bevels, chunky borders, inset wells, high-contrast Minecraft-like iconography.
- The screen is dense but organized. Main layout is a fixed top masthead, left sidebar, broad center column, and right utility rail. At desktop width, the center column is approximately 60% of remaining horizontal area, right rail approximately 25%, sidebar approximately 14% of full width.
- Bottom border is a continuous grass-and-dirt block strip. Lanterns, leaves, terrain and a scenic sidebar illustration are decorative, not UI controls.

### 2.2 Header / masthead (approximately y=0–124)

From left to right:

1. **Grass-block cube logo** at far left, visually about 80×80 px. Decorative brand mark; clicking it navigates to Overview.
2. **Two-line title:** `FAMILY GENERATIONAL` / `MINECRAFT SERVER`. Large pale gray beveled pixel-font text with shadow. Keep exact line break on desktop.
3. **Server status inset:** green status light, uppercase `ONLINE` (or OFFLINE / STARTING / STOPPING / UNKNOWN), with a small `Java Edition <version>` subtitle. This is real status data, never hard-coded.
4. **Right-end avatar and quote:** a Minecraft-style Steve face icon and text `Good friends build better worlds.` Decorative tagline, editable via configuration.
5. Background: black stone masonry, dark timber, vegetation, hanging lanterns with orange glow; thick horizontal separators.

### 2.3 Sidebar (approximately x=0–224, y=125–916)

A vertical stone-bordered navigation panel. Each nav item is a wide dark rectangular row about 60 px tall with an icon on the left and readable light text:

- **Overview** (house icon): default route `/`; green selected background, lighter green edge and inset highlight. Shows main dashboard.
- **Players** (Steve/player head): route `/players`; current and historical players, presence, player controls and optional profile details.
- **World** (grass block): route `/world`; world metadata, weather/time and optional map integration.
- **Console** (terminal prompt): route `/console`; streaming server log and privileged command interface.
- **Backups** (chest): route `/backups`; backup history, creation, retention, download and guarded restore.
- **Settings** (gear): route `/settings`; connectivity, appearance, server metadata, permissions, alert preferences and health checks.

Under navigation, fill the remaining sidebar with the screenshot's vertical pixel-art scene: a warm-lit wooden Minecraft cottage on a cliff, waterfall, lush foliage, sunset glow. It is decorative; maintain cropping rather than stretching. On narrow screens collapse navigation into an accessible drawer, retaining the same icons.

### 2.4 Main hero panel (approximately x=230–1144, y=129–484)

- Dominant panoramic Minecraft landscape: sunset over a blocky mountain valley with trees, rivers, a glowing village, and distant structures.
- Thick irregular dark stone frame, subtle inner shadow, clipped image corners. Preserve the screenshot's composition and sunset palette.
- Lower-left overlay: bright green Creeper face icon, huge beveled uppercase text **`WORLD ONLINE`**, smaller caption **`Adventure continues...`**.
- Overlay is part of the live state: `WORLD OFFLINE` and contextual caption when offline; do not falsely show online status.
- This panel is not a live Minecraft screenshot by default. It is a themed hero artwork; optionally allow replacing it with a user-provided world screenshot.
- Optional click action: navigate to World. No fake play button.

### 2.5 KPI strip (approximately x=230–1144, y=488–611)

Four equal-width inset stone cards, with pixel icons, small uppercase labels, prominent values, and colored horizontal progress indicators:

1. **PLAYERS ONLINE** — Steve head icon, e.g. `3 / 20`. Actual connected players and server max slots. Green bar = ratio of online to max. Clicking opens Players.
2. **TICK RATE (TPS)** — diamond pickaxe icon, e.g. `19.8 / 20`. Actual TPS when the Minecraft implementation exposes it reliably; bar max 20. If unavailable show `N/A`, not an invented value. TPS may require a server-compatible command/plugin; don't assume vanilla RCON exposes it.
3. **MEMORY USAGE** — grass block icon, e.g. `6.4 GB / 8 GB`. Distinguish JVM memory, container memory and host memory; default to measured Minecraft container usage / configured container limit. Amber bar when high; green at normal levels. Do not use fabricated readings.
4. **UPTIME** — clock icon, e.g. `2h 34m`. Real time since server/container start. The reference shows a green decorative progress bar; uptime has no natural percentage, so use a non-misleading activity bar rather than implying a maximum.

Each card has a dark beveled border, compact white numerals, subtle gradients, and a pixelated inset bar.

### 2.6 Lower main area (approximately x=230–1144, y=614–914)

Three adjacent cards:

#### A. WORLD INFORMATION (left; ~40% of lower area)

Header: grass-block icon and uppercase `WORLD INFORMATION`. Divider underneath. Compact labeled rows with small icons, left label and right-aligned value:

- **World Name** — actual level name, e.g. `family_world`.
- **Game Mode** — Survival / Creative / Adventure / Spectator, from server config or verified query.
- **Difficulty** — Peaceful / Easy / Normal / Hard, from config or verified query.
- **Seed** — only if server access and authorization permit; consider masking by default, with reveal control.
- **Dimension** — if no specific player is selected, label `Primary World: Overworld`; do not claim the entire server has one current dimension.
- **World Size** — disk usage of configured world directory, computed asynchronously and cached.
- **Last Backup** — time since last **verified successful** backup; if none show `Never`.

The sample values in the image are decorative and must be replaced with real data.

#### B. WORLD TIME & WEATHER (middle; ~32%)

Header: round clock icon and `WORLD TIME & WEATHER`. Upper portion is a two-column unit: a pixel-art sunset thumbnail and a dark inset reading Minecraft day number plus converted clock time. Lower weather box contains a sun icon, `Clear Skies` and explanatory text. Query time/weather when technically available, otherwise show `Unavailable`. Minecraft time conversion should reflect game ticks and be labeled **Minecraft Time**, not local system time. Weather may be inferred from appropriate command responses only when reliable; avoid guessing.

#### C. QUICK ACTIONS (right; ~28%)

Header: tool icon and `QUICK ACTIONS`. Five stacked full-width beveled buttons with pixel icons:

- **Start Server** — bright grass green. Start configured Docker container via restricted backend control; disabled while already running.
- **Stop Server** — dark red. Requires confirmation, shows connected-player count and warns of disconnection. Graceful shutdown; never `docker kill` as the default.
- **Restart Server** — amber/brown. Confirmation, optional player broadcast/countdown, graceful restart, visible progress.
- **Create Backup** — dark slate. Invoke safe backup job; show progress, result and failure detail. Prefer a consistent server-aware backup mechanism rather than blindly archiving actively changing files.
- **View Backups** — dark slate. Opens `/backups`.

No lifecycle button should perform a host operation without explicit backend authorization and status checks.

### 2.7 Right rail (approximately x=1150–1535, y=126–916)

Two stacked panels with narrow stone gutter.

#### A. ONLINE PLAYERS (top; ~350 px tall)

- Header with small green two-person pixel icon and text `ONLINE PLAYERS (N)`.
- Three example player rows in screenshot: `DeavaPrimus`, `RealBBQSauce`, `AlexTheBuilder`. **These are screenshot examples, not guaranteed active players.**
- Each row: pixelated Minecraft skin face/avatar on left, white username, row of red pixel hearts underneath, and green connection bars at far right.
- Actual online list from reliable server query. Avatars should use player skin/head assets with graceful default placeholder.
- **Hearts and connection bars must not be fabricated.** Health requires a compatible server-side data source, permission and a reliable per-player query; ping requires a compatible protocol or plugin. Hide unavailable metrics or replace with explicitly labeled unknown state.
- Bottom full-width `View All Players (N / max)` button navigates to `/players`.
- When no one is online, show `No players online` and a subdued decorative illustration, not fake player rows.

#### B. SERVER CONSOLE (below players; ~330 px tall)

- Header: terminal icon, `SERVER CONSOLE`, green live dot and `Live` when WebSocket is connected. On disconnection show `Reconnecting` / `Disconnected` clearly.
- Dark scrollable monospace log area, each line prefixed by time `[HH:MM:SS]`; color-code informational text, player joins, chat and server events as shown.
- Bottom entry field `Send a command...` with green square arrow/send button.
- **Enhancement agreed after the image:** add small **Console | Chat** tabs directly under this panel header, without changing its outer dimensions, placement or overall appearance. `Console` retains the screenshot's log/command view; `Chat` shows actual player messages and lets admins send messages into Minecraft. Remember selected tab per session.
- Console tab: commands go through RCON, with an explicit distinction between RCON commands and OS shell access. Do not expose host shell commands through this panel.
- Chat tab: scrollable timestamped chat with sender avatars/names, admin messages, send field and clear send feedback. New messages update live without page refresh.
- The Minecraft chat input is **not** a web-only conversation: messages must actually appear inside the game.

### 2.8 Footer

A visually continuous strip of green grass atop brown dirt blocks along the full bottom edge, approximately 35–45 px tall. Decorative only. Avoid covering interactive controls at smaller viewport heights.

## 3. Visual design system

### 3.1 Colors (approximate starting tokens, refine by sampling reference)

| Token | Suggested value | Usage |
|---|---|---|
| `--bg-obsidian` | `#0D1013` | Page background |
| `--stone-deep` | `#15191D` | Card interior |
| `--stone-panel` | `#20252A` | Inset surfaces |
| `--stone-edge` | `#3A4146` | Beveled outer edges |
| `--stone-shadow` | `#080A0C` | Deep frame shadows |
| `--grass` | `#48BA2F` | Active nav, positive bars |
| `--grass-bright` | `#63D83E` | Status LEDs, highlights |
| `--amber` | `#E5A126` | Memory warnings, restart |
| `--danger` | `#A92E29` | Stop, destructive actions |
| `--text-primary` | `#E5E4E2` | Headings and data |
| `--text-muted` | `#A4ADB6` | Labels and metadata |
| `--chat-green` | `#8FD37D` | Console events |

Colors are *approximations*, not image-extracted exact values. Pixel-accurate reconstruction requires visual comparison to the PNG.

### 3.2 Typography and icons

- Primary heading font: a licensed Minecraft-like block/pixel typeface with beveled/shadow treatment; avoid embedding copyrighted game fonts without license.
- Labels: readable pixel-style or squared monospace typeface. Prioritize readability at 12–16 px.
- Main title: very large, embossed, gray/white uppercase with dark outline and offset shadow.
- Numerals: crisp, legible, mostly white, with consistent baseline alignment.
- Icons: use coherent pixel-art sprite assets: grass block, creeper, Steve-style face, pickaxe, clock, chest, gear, terminal, sun, hearts, lantern. Do not mix generic thin-line Lucide icons into the primary themed design.
- Use nearest-neighbor (`image-rendering: pixelated`) on pixel sprites, **not** on scenic artwork or normal text.

### 3.3 Textures, borders and assets

- Implement reusable `StonePanel`, `StoneInset`, `PixelButton`, `PixelProgress`, `StatusLamp`, `PixelAvatar`, and `PanelHeader` components.
- Border design: 3–6 px stepped bevel with light top/left edges, dark bottom/right edges, deep inner shadow and slightly irregular stone texture.
- Background assets should be separated into hero landscape, sidebar cottage/waterfall, masthead stone/wood/lantern, and footer grass/dirt strip.
- The reference PNG is **not a sprite sheet**. Do not crop interface text and values from it and use those as functional UI. Recreate panels and text with HTML/CSS and separately sourced/generated/licensed art assets. If no exact assets exist, create closely matched original pixel-art assets.
- Avoid visual drift: do not replace pixel art with emoji, flat cards, modern gradients or rounded SaaS widgets.

### 3.4 Responsive rules

- **Desktop ≥1440 px:** maintain three-column screenshot composition and density.
- **Small desktop 1024–1439 px:** narrower sidebar and rail, maintain center hero; allow vertical scrolling.
- **Tablet 768–1023 px:** collapsible sidebar; hero + KPIs above two-column panels; right rail moves below main content.
- **Mobile <768 px:** single column; header compact, nav drawer, hero full width, KPI cards 2×2, world panels stacked, chat and console fully usable.
- Never hide safety confirmations or truncate commands silently. Provide visible focus states and sufficient contrast while preserving visual style.

## 4. Page-by-page functional specification

### 4.1 Overview `/`

All reference-image panels, live status, auto-updating KPIs, quick actions, player preview, console/chat tabs, responsive layout. Clicks route to appropriate detail screens. If data source is unavailable, show `Unavailable` or `Unknown`, not zero unless verified.

### 4.2 Players `/players`

- Online players list and count; offline/known players when data is available.
- Player detail: username, UUID, skin, first/last seen, session duration, online state and historical activity where recorded.
- Admin actions (subject to role and capability): message player, broadcast, whitelist add/remove, op/deop with strong confirmation, kick, ban/unban, teleport or gamemode changes where permitted.
- Player health, XP, inventory and precise position are **optional advanced capabilities**, requiring supported commands/plugins and verification; don't pretend RCON alone guarantees every field.
- Record all administrative actions in an audit log.

### 4.3 World `/world`

- World name, seed visibility controls, game mode, difficulty, size, version, time/weather and world folder information.
- Optional interactive map only after installing/configuring a compatible map renderer such as BlueMap or Dynmap; never claim a map is available by default.
- Server-approved time/weather adjustments with confirmation and audit events.
- Clear separation between static configuration and real-time observations.

### 4.4 Console `/console`

- Full-screen searchable streaming log view, pause/resume auto-scroll, log level filters, timestamps, connection status, command input and command response history.
- Command execution via authenticated, authorized RCON adapter only.
- Mask secrets and sensitive environment variables; no unrestricted host shell.
- Commands and responses stored with configurable retention and audit metadata.

### 4.5 Chat (Overview tab and optionally `/chat`)

**Core required feature: true two-way communication with Minecraft players.**

- Read new in-game player chat messages from Minecraft logs or a compatible event/plugin bridge; parse sender, message and timestamp. Handle log rotation and reconnect.
- Display chat as it happens via WebSocket. Preserve ordering, deduplicate after reconnection, persist to a small SQLite database.
- Send dashboard messages into Minecraft using RCON `tellraw` or another version-compatible server command. Display sender clearly as `Dashboard Admin` or configured label.
- **Never concatenate untrusted input directly into raw commands.** Properly encode tellraw JSON, validate length, enforce rate limits and escape control characters.
- Support global chat initially; direct/private messages only if command and player targeting are verified.
- Distinguish outgoing delivery acknowledgment from actual in-game receipt. Show send failures.
- Separate chat from server log messages; do not accidentally treat join/leave/death notifications as player chat.
- Chat history should be configurable for privacy and retention.
- Players require no mod/plugin for ordinary global chat on a compatible vanilla server; richer events may require server extensions.

### 4.6 Backups `/backups`

- Backup list with creation time, size, status, verification and source world.
- On-demand backup, configurable schedule and retention.
- Prefer an atomic/safe backup strategy appropriate to the server and storage; for a running Minecraft world, coordinate saves (`save-off` / `save-all flush` / `save-on`) if supported, with `finally` cleanup so saving is re-enabled after errors.
- Never delete world files or overwrite the active world as part of backup creation.
- Restore requires a separate, explicit workflow: verify backup, warn about downtime, stop server safely, create pre-restore safety copy, restore and restart; ideally test restoration in a separate environment first.
- Show integrity/checksum information and failed job details.

### 4.7 Settings `/settings`

- Server display name, banner/tagline, image assets, refresh cadence, theme options, timezone.
- Read-only connectivity diagnostics for Docker, RCON, log file and world directory.
- Credentials stored in environment/secret files, not exposed in UI.
- Authentication, session management, user roles, backup retention and audit history.
- Allow read-only mode when lifecycle management is unavailable.

## 5. Backend integration and technical constraints

### 5.1 Recommended architecture

| Layer | Technology | Responsibility |
|---|---|---|
| Frontend | React + TypeScript + Vite | Pixel-themed responsive interface |
| Styling | Tailwind CSS + custom CSS/texture assets | Precise visual recreation |
| Backend | Go | API, authorization, Minecraft integration |
| Real time | WebSockets | Chat, logs, status and jobs |
| Persistence | SQLite | Chat history, audit, sessions, backup metadata |
| Minecraft | Existing `itzg/minecraft-server` | Game server; do not replace |
| Command bridge | Minecraft RCON | Controlled in-game commands and broadcasts |
| Lifecycle | Restricted Docker control adapter | Start/stop/restart container |
| Deployment | Docker Compose | Self-hosted service deployment |

### 5.2 Required integration adapters

1. **Status:** Minecraft server query/ping and Docker health where available. Distinguish `game reachable`, `container running`, and `RCON reachable`.
2. **RCON:** TCP client to existing server's internal RCON port; keep RCON off the public internet. Add reconnect, timeout and command allowlisting by role.
3. **Logs:** Read-only mounted Minecraft `logs/latest.log` with inode/rotation detection, initial tail, parsing, and resilient reconnect.
4. **Docker lifecycle:** Do not mount unrestricted `/var/run/docker.sock` into a public-facing dashboard. Use a dedicated least-privilege local helper with only allowlisted actions on the one Minecraft container, or another tightly scoped orchestration method.
5. **Metrics:** Container memory/CPU from a trusted metrics adapter; TPS only from supported server capability; uptime from actual container/server lifecycle.
6. **Files:** Read-only world/config metadata by default; backups through dedicated writable backup destination with constrained paths.
7. **Optional plugins:** World map, inventory, health and advanced telemetry should be feature-gated and documented as requiring extra integration.

### 5.3 Suggested REST endpoints (implementation contract, not existing APIs)

```text
GET    /api/v1/status
GET    /api/v1/metrics
GET    /api/v1/players
GET    /api/v1/players/{uuid}
GET    /api/v1/world
GET    /api/v1/logs?cursor=...
GET    /api/v1/chat?before=...&limit=...
POST   /api/v1/chat/messages
POST   /api/v1/console/commands
POST   /api/v1/server/start
POST   /api/v1/server/stop
POST   /api/v1/server/restart
GET    /api/v1/backups
POST   /api/v1/backups
POST   /api/v1/backups/{id}/restore
GET    /api/v1/jobs/{id}
GET    /api/v1/settings
PATCH  /api/v1/settings
GET    /api/v1/capabilities
GET    /api/v1/audit
GET    /api/v1/health
WS     /api/v1/events
```

All mutations must be authenticated, role-checked, validated, audited and protected against CSRF when using cookie-based sessions. Destructive operations require a distinct confirmation step. APIs should return structured errors, stable IDs and capability flags.

### 5.4 WebSocket event types

`server.status`, `server.metrics`, `player.join`, `player.leave`, `chat.message`, `console.line`, `backup.progress`, `backup.completed`, `backup.failed`, `server.operation.progress`, `connection.state`.

Each event should carry an ID, timestamp, type, server ID and typed payload. Reconnect should support replay or cursor-based history to avoid gaps/duplicates.

### 5.5 Security

- Private deployment preferred via existing Cloudflare Zero Trust, VPN or authenticated reverse proxy; never publish RCON.
- Authentication mandatory even on LAN. Admin and viewer roles; only admin can send chat, run commands, control server or restore backups.
- Argon2id password hashes if implementing local accounts; secure HttpOnly SameSite cookies, CSRF protection, rate limiting and session expiry.
- Strict validation for command/chat inputs. Separate chat-send permission from arbitrary RCON command permission.
- No arbitrary filesystem browser, no unrestricted Docker socket, no OS shell through the console.
- Audit: login, command, message send, backup, restore, stop/start/restart, permission change.
- Bind services to appropriate private interfaces; secrets through environment/secret mounts, never committed to Git.

## 6. Docker Compose deployment expectations

- Add a new dashboard stack without modifying existing Minecraft world volumes or replacing the Minecraft container.
- Prefer dashboard backend/frontend as one or two containers, SQLite stored in a persistent bind-mounted local directory, and dedicated backup directory.
- Mount Minecraft logs and relevant configuration read-only; use narrowly scoped separate control mechanisms for RCON and Docker actions.
- If dashboard and Minecraft run in separate Compose projects, connect them through an explicit private network or configured host address; do not assume Docker DNS names resolve across projects.
- Configuration variables should cover Minecraft host, game port, RCON host/port/password, container identifier, log path, world path, backup path, public URL, session secret and timezone.
- Provide `.env.example`, `compose.yaml`, README, backup/restore instructions and a health endpoint.
- **Never include a real RCON password or secret in examples.**

## 7. Exact behavioral expectations

| UI element | Click / update behavior |
|---|---|
| Grass-block masthead logo | Navigate to Overview |
| Online badge | Show current status and tooltip/details for game, Docker and RCON connectivity |
| Sidebar rows | Navigate to corresponding pages and update selected state |
| Hero panel | Show world/server state; optional navigation to World |
| Players KPI | Open Players |
| TPS KPI | Show source/capability and latest measurement timestamp |
| Memory KPI | Show measured memory and limit, not host RAM mislabeled as JVM RAM |
| Uptime KPI | Update from actual start time |
| Player row | Open player detail |
| View All Players | Navigate to Players |
| Console/Chat tabs | Switch within same right-rail panel |
| Console send | Execute authorized RCON command and show response/error |
| Chat send | Broadcast correctly formatted in-game message, display status |
| Start / Stop / Restart | Guarded backend operation with progress and confirmations |
| Create Backup | Create job and show live progress |
| View Backups | Navigate to backup management |
| World time/weather | Update from supported data source; show unavailable otherwise |

## 8. State, loading and error handling

- On initial load, show themed skeleton panels while fetching data. Do not briefly display sample mock values.
- Distinguish: server offline, backend offline, RCON unavailable, insufficient permission, metric unsupported, log stream disconnected and backup failure.
- Show time of last successful data refresh and stale-data indication.
- Disable incompatible controls based on capability and state (e.g. Start when online, Stop when offline).
- Show accessible confirmation dialogs for disruptive actions; announce errors and success in UI.
- WebSocket reconnect uses bounded exponential backoff and shows a connection badge.
- Log/chat views retain scroll position if user is reading history; only auto-scroll if already at bottom.

## 9. Implementation phases

1. **Visual foundation:** Reproduce screenshot geometry, pixel assets, masthead, sidebar, hero, KPIs, lower panels and right rail with fixtures. Review side-by-side at 1536×960.
2. **Read-only live integration:** Status, players, memory, uptime, logs, world metadata; replace fixtures with real data; show unknown for unsupported metrics.
3. **Chat and console:** Two-way in-game chat, console commands, WebSocket delivery, history and audit.
4. **Management:** Safe start/stop/restart, backup creation/listing and guarded restore.
5. **Detail pages:** Players, World, Console, Backups and Settings using the same visual language.
6. **Hardening:** Authentication, authorization, CSRF, rate limits, restricted Docker access, tests, documentation, deployment and recovery testing.

## 10. Acceptance criteria / definition of done

### Visual

- [ ] At **1536×960**, screenshot comparison shows the same three-column layout, header proportions, hero position, four KPI cards, lower three cards, right-side player/console stack, stone bevels, sidebar art and grass footer.
- [ ] No generic modern SaaS cards, emoji icon substitutions, excessive rounding or glass effects.
- [ ] Pixel art and scenic assets look consistent with the reference; labels and values are real HTML text.
- [ ] The **Console | Chat** tabs fit within the existing console panel without materially changing its silhouette.
- [ ] Desktop, tablet and mobile layouts remain usable without clipping interactive elements.

### Functional

- [ ] Online/offline status matches actual Minecraft connectivity.
- [ ] Online player list reflects the game; unsupported health/ping indicators are not fabricated.
- [ ] Real-time in-game player chat appears in dashboard within a few seconds under normal connectivity.
- [ ] Dashboard messages appear in Minecraft game chat, with correctly escaped content.
- [ ] Console executes permitted RCON commands and returns result/errors.
- [ ] Server lifecycle controls are confirmed, audited and restricted to the configured Minecraft container.
- [ ] Backups are safely created and can be verified; restore is guarded.
- [ ] Server remains fully functional and existing world data remains intact after dashboard installation.
- [ ] No sensitive ports or credentials are publicly exposed.
- [ ] All unknown/unsupported metrics display honest fallback states.

### Documentation and tests

- [ ] README includes architecture, configuration, local development, deployment, rollback and troubleshooting.
- [ ] Integration tests cover RCON send/timeout, log parsing/rotation, chat injection escaping, WebSocket reconnect and backup failures.
- [ ] Authorization tests ensure viewer cannot issue commands, chat messages or lifecycle actions.
- [ ] E2E tests cover dashboard navigation, chat, console, confirmations and failure states.

## 11. Instructions for the implementing AI/developer

**Read this entire Markdown file and inspect `minecraft-dashboard-reference.png` before writing UI code.** Treat the image as the visual design target, and this file as the functional specification. Do not redesign the interface or substitute a generic admin theme. Build reusable components and preserve pixel fidelity. Use genuine backend integrations, not mock values in production. Do not assume that vanilla Minecraft RCON exposes TPS, inventory, health, ping, world maps or all metadata. Identify unsupported capabilities explicitly. Prioritize a safe, additive deployment to the existing `itzg/minecraft-server` setup. Ask only for configuration values that cannot be safely discovered, such as RCON credentials, Docker control approach and actual volume locations. Never invent them. Deliver the running application, source, deployment files, tests and documentation.

## 12. Reference-image inventory (exact visible copy)

Visible text in the supplied illustration includes: `FAMILY GENERATIONAL MINECRAFT SERVER`, `ONLINE`, `Java Edition 1.20.4`, `Good friends build better worlds.`, `Overview`, `Players`, `World`, `Console`, `Backups`, `Settings`, `WORLD ONLINE`, `Adventure continues...`, `PLAYERS ONLINE`, `TICK RATE (TPS)`, `MEMORY USAGE`, `UPTIME`, `WORLD INFORMATION`, `World Name`, `Game Mode`, `Difficulty`, `Seed`, `Dimension`, `World Size`, `Last Backup`, `WORLD TIME & WEATHER`, `QUICK ACTIONS`, `Start Server`, `Stop Server`, `Restart Server`, `Create Backup`, `View Backups`, `ONLINE PLAYERS (3)`, `View All Players (3 / 20)`, `SERVER CONSOLE`, `Live`, `Send a command...`.

These exact labels guide visual reconstruction. Illustrated numeric values, Minecraft version, world name, player names and sample logs are **examples only** and must be dynamically populated or correctly marked unavailable.
