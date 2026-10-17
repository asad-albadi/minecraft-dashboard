package capabilities

import "minecraft-dashboard/backend/internal/config"

type Snapshot struct { Chat bool `json:"chat"`; Console bool `json:"console"`; Lifecycle bool `json:"lifecycle"`; Backups bool `json:"backups"`; Reason string `json:"reason,omitempty"` }
func SnapshotFor(c config.Config) Snapshot { rcon := c.RCONAddress != "" && c.RCONPassword != ""; reason := "Configure read-only paths and explicit RCON settings to enable integrations."; if rcon { reason = "Live read-only RCON status is connected. Chat, console, lifecycle, and backups remain disabled until authenticated, least-privilege adapters are installed." }; return Snapshot{Chat:false, Console:false, Lifecycle:false, Backups:false, Reason:reason} }
