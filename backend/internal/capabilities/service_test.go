package capabilities

import (
	"minecraft-dashboard/backend/internal/config"
	"testing"
)

func TestReadOnlyRCONDoesNotEnableMutations(t *testing.T) {
	snapshot := SnapshotFor(config.Config{RCONAddress: "mc:25575", RCONPassword: "secret", ContainerName: "mc"})
	if snapshot.Chat || snapshot.Console || snapshot.Lifecycle || snapshot.Backups {
		t.Fatalf("read-only RCON must not enable a mutation capability: %#v", snapshot)
	}
}

func TestActionsRequireExplicitConfiguration(t *testing.T) {
	c := config.Config{RCONAddress: "mc:25575", RCONPassword: "secret", AccountsFile: "/private/accounts.json", SessionSecret: "12345678901234567890123456789012", DataPath: "/dashboard-data", EnableChat: true, EnableConsole: true}
	s := SnapshotFor(c)
	if !s.Chat || !s.Console || s.Lifecycle || s.Backups {
		t.Fatalf("capabilities=%+v", s)
	}
	c.DataPath = ""
	s = SnapshotFor(c)
	if s.Chat || s.Console {
		t.Fatal("actions enabled without audit storage")
	}
}
