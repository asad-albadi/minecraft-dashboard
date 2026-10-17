package capabilities

import "minecraft-dashboard/backend/internal/config"

type Snapshot struct { Chat bool `json:"chat"`; Console bool `json:"console"`; Lifecycle bool `json:"lifecycle"`; Backups bool `json:"backups"`; Reason string `json:"reason,omitempty"` }
func SnapshotFor(c config.Config) Snapshot { rcon := c.RCONAddress != "" && c.RCONPassword != ""; backup := c.BackupPath != "" && c.WorldPath != "" && rcon; lifecycle := c.ContainerName != "" && rcon; reason := "Configure read-only paths and explicit RCON/control settings to enable integrations."; if rcon { reason = "Lifecycle and backups remain disabled until their explicit paths/container are configured." }; return Snapshot{Chat:rcon, Console:rcon, Lifecycle:lifecycle, Backups:backup, Reason:reason} }
