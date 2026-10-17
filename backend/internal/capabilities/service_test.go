package capabilities

import (
	"testing"
	"minecraft-dashboard/backend/internal/config"
)

func TestReadOnlyRCONDoesNotEnableMutations(t *testing.T) {
	snapshot := SnapshotFor(config.Config{RCONAddress:"mc:25575", RCONPassword:"secret", ContainerName:"mc"})
	if snapshot.Chat || snapshot.Console || snapshot.Lifecycle || snapshot.Backups { t.Fatalf("read-only RCON must not enable a mutation capability: %#v", snapshot) }
}
