# Changelog

All notable changes to this project are recorded here.

## [Unreleased]

### Added

- Added a read-only Players view sourced from live vanilla player cache, operator, and world-stat files, with play time, deaths, distance walked, mob kills, operator marker, and skin heads.
- Replaced generic KPI glyph boxes with compact player, grass-block, crossed-tool, and difficulty indicators; refined the original pixel wordmark with a stone/iron texture.
- Added the verified Minecraft server version from the read-only startup log to the online status badge.
- Added per-player Minecraft skin-head rendering from the public player name, with an in-dashboard pixel-face fallback when the image is unavailable.

### Fixed

- Contained the desktop dashboard canvas within the viewport; only the server-console log now scrolls when content exceeds its panel.
- Added dark carved-stone texture to unused desktop canvas space.
- Added a read-only Server-Sent Events (SSE) stream for live server status and sanitized console-log additions, with automatic browser reconnection.
- Original cinematic voxel hero-valley and cliffside-cottage artwork, integrated into the dashboard’s hero and decorative sidebar.
- Configured the local dashboard environment with existing Minecraft read-only log/world paths and private network discovery.
- Added verified world metadata: configured world name, game mode, difficulty, masked seed state, primary dimension, and asynchronously calculated world size.
- Added a read-only initial Minecraft log tail to the console panel, with IP-address masking.
- Added conservative vanilla-RCON parsing for Minecraft daytime and weather; unsupported command responses remain explicitly unavailable.
- Tightened the desktop reference layout so the lower dashboard cards flex within the center column and long player/log text wraps or truncates without crossing into the utility rail.
- Reworked the dashboard’s visual primitives with original stepped voxel blockwork, lantern glow, grass-cube branding, pixel-face avatars, inset console texture, and heavier stone bevels to more closely match the supplied visual reference.
- Matched the 1536×960 reference composition more closely with a measured 350 px utility rail, expanded center canvas, fixed-height desktop shell, non-wrapping panel headers, and a visible in-frame footer.
- Corrected the voxel wordmark palette and removed fixed-height desktop clipping so constrained-height screens scroll and reflow naturally.
- Kept the desktop reference composition as a large-screen calibration while allowing normal document scrolling at constrained heights.
- Replaced unsupported TPS, memory, uptime, seed, and weather placeholders with verified world metadata; added live Minecraft time for the server's current vanilla clock syntax.
- Bun/Vite/React pixel-themed dashboard shell matching the supplied desktop composition and adapting for tablet/mobile.
- Go/Fiber API skeleton with health, capability, read-only placeholder data, and safety-gated mutation routes.
- Safe Minecraft `tellraw` JSON command construction with tests for escaping and control-character rejection.
- Additive Docker Compose deployment, private localhost port bindings, and container build definitions.
- Project contributor instructions, operational documentation, and troubleshooting guidance.

### Security

- Disabled all mutating controls while the dashboard has only read-only RCON integration.
- Server management capabilities remain disabled until explicit private integration settings are provided.
- No Minecraft server volume, player data, RCON secret, or Docker socket is included in deployment defaults.
