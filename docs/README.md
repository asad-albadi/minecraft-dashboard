# Documentation

- [Setup and account security](setup.md)
- [Backup and restore boundaries](backups.md)
- [Implemented features and remaining integrations](feature-coverage.md)
- [Current implementation plan](implementation-plan.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [Contributor rules](../AGENTS.md)
- [Product vision](../MINECRAFT_DASHBOARD_SPEC.md)

The dashboard uses authenticated same-origin REST and SSE. Its services own only dashboard account/session/history/audit state. Minecraft paths remain read-only; no Docker socket or host shell is exposed.
