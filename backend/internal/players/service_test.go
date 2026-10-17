package players

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadProfilesCombinesLiveCacheOperatorsAndWorldStats(t *testing.T) {
	dir := t.TempDir()
	cache := filepath.Join(dir, "usercache.json")
	ops := filepath.Join(dir, "ops.json")
	world := filepath.Join(dir, "world")
	stats := filepath.Join(world, "players", "stats")
	if err := os.MkdirAll(stats, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, []byte(`[{"uuid":"alpha","name":"Alex"},{"uuid":"beta","name":"Bee"}]`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ops, []byte(`[{"uuid":"alpha","name":"Alex","level":4}]`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stats, "alpha.json"), []byte(`{"stats":{"minecraft:custom":{"minecraft:play_time":144000,"minecraft:deaths":3,"minecraft:walk_one_cm":1234500,"minecraft:mob_kills":12}}}`), 0600); err != nil {
		t.Fatal(err)
	}

	profiles, err := ReadProfiles(cache, ops, world)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("profiles = %#v", profiles)
	}
	profile := profiles[0]
	if profile.Name != "Alex" || !profile.Operator || profile.PlayTime != "2.0 h" || profile.Deaths != "3" || profile.Walked != "12.3 km" || profile.MobKills != "12" {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}
