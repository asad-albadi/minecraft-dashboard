# Feature coverage and remaining integrations

This is the implementation record, not a claim that every item in the product vision is finished.

| Area | Implemented | Remaining / intentionally unavailable |
| --- | --- | --- |
| Identity | Generic configurable server title; named admin/viewer accounts; local preferences | Web account provisioning / password changes |
| Overview | Java status ping separate from RCON, version/count/latency when available, world facts, online players, recent activity, navigation/refresh/copy | CPU, memory, network, uptime, TPS/MSPT require a restricted reliable metrics source |
| Players | Directory, name search, presence/operator/recent filters, sorting, detailed raw vanilla stats, saved advancements, retained join dates/counts | Reliable inventory/armor/offhand icons, complete sessions/deaths/activity timeline, direct profile links from online rows, player administration |
| World | Cached sizes/dimensions, configured mode/difficulty/hardcore/generator, sampled time/day/weather when supported | NBT spawn/age metadata beyond RCON day, actual loaded dimensions, external read-only map renderer, world controls |
| Console | Bounded streaming logs, search/severity/pause/copy/export/clear, explicit opt-in read-only commands and response history | Arbitrary or high-risk commands deliberately forbidden; WebSocket transport not implemented (SSE retained) |
| Chat/activity | Recognized vanilla chat/events, persisted bounded histories, optional audited outgoing chat | Custom/localized log adapters, full replay across server/dashboard downtime, configurable retention/SQLite migration |
| Backups | Read-only archive catalog with honest unknown integrity | Coordinated creation, downloads, checksums from trusted manifests, deletion, retention jobs, restore; require separate safe workflows |
| Settings | Theme/density/timezone/clock/refresh/landing/scroll/notice preferences; account role/sign-out; non-secret connection information | Browser settings are local, not account-synced; notifications are in-app connection notices, not OS push |

All Minecraft-changing features are off by default. Chat opt-in is the only implemented game-changing operation. Lifecycle and restore cannot be enabled through environment flags. No Minecraft container, world, properties or player data is modified during dashboard installation or testing.
