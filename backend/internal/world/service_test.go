package world

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMetadataUsesVerifiedServerPropertiesAndWorldSize(t *testing.T) {
	dir := t.TempDir()
	properties := filepath.Join(dir, "server.properties")
	if err := os.WriteFile(properties, []byte("level-name=family_world\ngamemode=survival\ndifficulty=easy\nlevel-seed=12345\n"), 0600); err != nil { t.Fatal(err) }
	worldDir := filepath.Join(dir, "world")
	if err := os.Mkdir(worldDir, 0755); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(worldDir, "region.mca"), make([]byte, 1536), 0600); err != nil { t.Fatal(err) }
	metadata, err := ReadMetadata(properties, worldDir)
	if err != nil { t.Fatal(err) }
	if metadata.Name != "family_world" || metadata.GameMode != "Survival" || metadata.Difficulty != "Easy" || metadata.Size != "1.5 KB" { t.Fatalf("unexpected metadata: %#v", metadata) }
}
