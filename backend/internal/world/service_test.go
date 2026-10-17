package world

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMetadataUsesVerifiedServerPropertiesAndWorldSize(t *testing.T) {
	dir := t.TempDir()
	properties := filepath.Join(dir, "server.properties")
	if err := os.WriteFile(properties, []byte("level-name=family_world\ngamemode=survival\ndifficulty=easy\nlevel-seed=12345\n"), 0600); err != nil {
		t.Fatal(err)
	}
	worldDir := filepath.Join(dir, "world")
	if err := os.Mkdir(worldDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worldDir, "region.mca"), make([]byte, 1536), 0600); err != nil {
		t.Fatal(err)
	}
	metadata, err := ReadMetadata(properties, worldDir)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Name != "family_world" || metadata.GameMode != "Survival" || metadata.Difficulty != "Easy" || metadata.Size != "1.5 KB" {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
}

func TestParseMinecraftTime(t *testing.T) {
	day, clock, ok := ParseMinecraftTime("The time is: 18475")
	if !ok || day != "Day 0" || clock != "00:28" {
		t.Fatalf("unexpected parsed time: day=%q clock=%q ok=%v", day, clock, ok)
	}
}

func TestParseMinecraftTimelineTime(t *testing.T) {
	day, clock, ok := ParseMinecraftTime("Timeline minecraft:day is at 9107 tick(s)")
	if !ok || day != "" || clock != "15:06" {
		t.Fatalf("unexpected timeline time: day=%q clock=%q ok=%v", day, clock, ok)
	}
}

func TestParseWeather(t *testing.T) {
	weather, ok := ParseWeather("The weather is: clear")
	if !ok || weather != "Clear Skies" {
		t.Fatalf("unexpected weather: %q ok=%v", weather, ok)
	}
}
