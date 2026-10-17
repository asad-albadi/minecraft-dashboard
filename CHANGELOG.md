# Changelog

All notable changes to this project are recorded here.

## [Unreleased]

### Added

- Original cinematic voxel hero-valley and cliffside-cottage artwork, integrated into the dashboard’s hero and decorative sidebar.
- Bun/Vite/React pixel-themed dashboard shell matching the supplied desktop composition and adapting for tablet/mobile.
- Go/Fiber API skeleton with health, capability, read-only placeholder data, and safety-gated mutation routes.
- Safe Minecraft `tellraw` JSON command construction with tests for escaping and control-character rejection.
- Additive Docker Compose deployment, private localhost port bindings, and container build definitions.
- Project contributor instructions, operational documentation, and troubleshooting guidance.

### Security

- Server management capabilities remain disabled until explicit private integration settings are provided.
- No Minecraft server volume, player data, RCON secret, or Docker socket is included in deployment defaults.
