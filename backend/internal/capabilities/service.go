package capabilities

import "minecraft-dashboard/backend/internal/config"

type Snapshot struct {
	Chat      bool   `json:"chat"`
	Console   bool   `json:"console"`
	Lifecycle bool   `json:"lifecycle"`
	Backups   bool   `json:"backups"`
	Reason    string `json:"reason,omitempty"`
}

func SnapshotFor(c config.Config) Snapshot {
	ready := c.RCONAddress != "" && c.RCONPassword != "" && c.DataPath != "" && c.AccountsFile != "" && len(c.SessionSecret) >= 32
	return Snapshot{Chat: ready && c.EnableChat, Console: ready && c.EnableConsole, Reason: "Chat and restricted read-only console require explicit enablement, administrator access, private RCON, and writable dashboard audit storage. Lifecycle and backups remain unavailable."}
}
