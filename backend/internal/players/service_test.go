package players

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadProfilesCombinesLiveCacheOperatorsAndWorldStats(t *testing.T) {
	dir := t.TempDir()
	cache := filepath.Join(dir, "usercache.json")
	ops := filepath.Join(dir, "ops.json")
	world := filepath.Join(dir, "world")
	stats := filepath.Join(world, "players", "stats")
	logs := filepath.Join(dir, "logs")
	if err := os.MkdirAll(stats, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(logs, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, []byte(`[{"uuid":"00000000-0000-0000-0000-000000000001","name":"Alex"},{"uuid":"00000000-0000-0000-0000-000000000002","name":"Bee"}]`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ops, []byte(`[{"uuid":"00000000-0000-0000-0000-000000000001","name":"Alex","level":4}]`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stats, "00000000-0000-0000-0000-000000000001.json"), []byte(`{"stats":{"minecraft:custom":{"minecraft:play_time":144000,"minecraft:deaths":3,"minecraft:walk_one_cm":1234500,"minecraft:mob_kills":12}}}`), 0600); err != nil {
		t.Fatal(err)
	}

	writeCompressedLog(t, filepath.Join(logs, "2026-10-01-1.log.gz"), "[12:00:00] [Server thread/INFO]: Alex joined the game\n")
	writeCompressedLog(t, filepath.Join(logs, "2026-10-03-1.log.gz"), "[12:00:00] [Server thread/INFO]: Alex joined the game\n")
	profiles, err := ReadProfiles(cache, ops, world, logs)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 || profiles[1].Deaths != "Unknown" {
		t.Fatalf("profiles = %#v", profiles)
	}
	profile := profiles[0]
	if profile.Name != "Alex" || !profile.Operator || profile.PlayTime != "2.0 h" || profile.Deaths != "3" || profile.Walked != "12.3 km" || profile.MobKills != "12" || profile.JoinEvents != "2" || profile.FirstJoin != "2026-10-01" || profile.LastJoin != "2026-10-03" {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestVanillaDetailsAndSafeFallbacks(t *testing.T) {
	dir := t.TempDir()
	uuid := "00000000-0000-0000-0000-000000000001"
	write := func(path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cache := filepath.Join(dir, "usercache.json")
	write(cache, `[{"uuid":"`+uuid+`","name":"Alex"},{"uuid":"../escape","name":"Unsafe"}]`)
	write(filepath.Join(dir, "stats", uuid+".json"), `{"stats":{"minecraft:mined":{"minecraft:stone":12},"minecraft:custom":{"minecraft:deaths":0,"minecraft:player_kills":2}}}`)
	write(filepath.Join(dir, "advancements", uuid+".json"), `{"DataVersion":3955,"minecraft:story/mine_stone":{"done":true,"criteria":{"get_stone":"2026-10-01 12:00:00 +0400"}},"minecraft:adventure/kill_all_mobs":{"done":false,"criteria":{"minecraft:zombie":"2026-10-02 12:00:00 +0400"}}}`)
	profiles, err := ReadProfiles(cache, filepath.Join(dir, "ops.json"), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(profiles)
	var decoded []map[string]any
	json.Unmarshal(data, &decoded)
	if len(decoded) != 2 {
		t.Fatalf("profiles: %s", data)
	}
	var p map[string]any
	for _, candidate := range decoded {
		if candidate["name"] == "Alex" {
			p = candidate
		}
	}
	if p["uuid"] != uuid || p["deaths"] != "0" || p["mobKills"] != "Unknown" || p["inventory"] != "Unavailable" || p["stats"] == nil {
		t.Fatalf("details: %s", data)
	}
	a, ok := p["advancements"].([]any)
	if !ok || len(a) != 2 {
		t.Fatalf("advancements: %s", data)
	}
	if a[0].(map[string]any)["progress"] != "Unknown" {
		t.Fatalf("incomplete progress must not invent denominator: %s", data)
	}
}

func TestStatsSymlinkEscapeIsUnavailable(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	uuid := "00000000-0000-0000-0000-000000000001"
	os.WriteFile(filepath.Join(dir, "cache.json"), []byte(`[{"uuid":"`+uuid+`","name":"Alex"}]`), 0600)
	os.WriteFile(filepath.Join(outside, uuid+".json"), []byte(`{"stats":{"minecraft:custom":{"minecraft:deaths":99}}}`), 0600)
	os.Symlink(outside, filepath.Join(dir, "stats"))
	profiles, err := ReadProfiles(filepath.Join(dir, "cache.json"), filepath.Join(dir, "ops.json"), dir, "")
	if err != nil || len(profiles) != 1 || profiles[0].Deaths != "Unknown" {
		t.Fatalf("unsafe source read: %#v %v", profiles, err)
	}
}

func TestJSONSizeLimitAndMalformedStatsStayUnknown(t *testing.T) {
	dir := t.TempDir()
	uuid := "00000000-0000-0000-0000-000000000001"
	os.Mkdir(filepath.Join(dir, "stats"), 0755)
	os.WriteFile(filepath.Join(dir, "cache.json"), []byte(`[{"uuid":"`+uuid+`","name":"Alex"}]`), 0600)
	for _, body := range []string{`{"stats":`, strings.Repeat(" ", 4*1024*1024+1)} {
		os.WriteFile(filepath.Join(dir, "stats", uuid+".json"), []byte(body), 0600)
		profiles, err := ReadProfiles(filepath.Join(dir, "cache.json"), filepath.Join(dir, "ops.json"), dir, "")
		if err != nil || len(profiles) != 1 || profiles[0].StatsStatus != "Unavailable" || profiles[0].Deaths != "Unknown" {
			t.Fatalf("malformed stats: %#v %v", profiles, err)
		}
	}
}

func TestLogHistoryDoesNotCountPlayerChatAsJoin(t *testing.T) {
	history := map[string]joins{}
	err := recordJoins(strings.NewReader("[12:00:00] [Server thread/INFO]: <Bob> Alex joined the game\n[12:01:00] [Server thread/INFO]: Alex joined the game\n"), "2026-10-01", history)
	if err != nil || history["Alex"].count != 1 {
		t.Fatalf("chat treated as join: %#v %v", history, err)
	}
}

func TestPartialHistoryPreservesUsableArchivesAndLabelsIncomplete(t *testing.T) {
	dir := t.TempDir()
	logs := filepath.Join(dir, "logs")
	os.Mkdir(logs, 0755)
	os.WriteFile(filepath.Join(dir, "cache.json"), []byte(`[{"uuid":"00000000-0000-0000-0000-000000000001","name":"Alex"}]`), 0600)
	writeCompressedLog(t, filepath.Join(logs, "2026-10-01-1.log.gz"), "[12:00:00] [Server thread/INFO]: Alex joined the game\n")
	os.WriteFile(filepath.Join(logs, "2026-10-02-1.log.gz"), []byte("damaged gzip"), 0600)
	writeCompressedLog(t, filepath.Join(logs, "2026-10-03-1.log.gz"), "[12:00:00] [Server thread/INFO]: Alex joined the game\n")
	os.WriteFile(filepath.Join(logs, "latest.log"), []byte(strings.Repeat("uninteresting log line\n", 200000)), 0600)
	profiles, err := ReadProfiles(filepath.Join(dir, "cache.json"), filepath.Join(dir, "ops.json"), dir, logs)
	if err != nil || len(profiles) != 1 {
		t.Fatalf("profiles: %#v %v", profiles, err)
	}
	p := profiles[0]
	if p.JoinEvents != "2" || p.FirstJoin != "2026-10-01" || p.LastJoin != "2026-10-03" || !strings.Contains(strings.ToLower(p.HistoryNote), "incomplete") {
		t.Fatalf("lost partial observations: %#v", p)
	}
}

func writeCompressedLog(t *testing.T, path, body string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := gzip.NewWriter(file)
	if _, err := writer.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
