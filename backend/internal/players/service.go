// Package players reads vanilla player state from read-only Minecraft data.
package players

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Profile struct {
	Name      string `json:"name"`
	Operator  bool   `json:"operator"`
	PlayTime  string `json:"playTime"`
	Deaths    string `json:"deaths"`
	Walked    string `json:"walked"`
	MobKills  string `json:"mobKills"`
	playTicks float64
}
type cacheEntry struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
type operatorEntry struct {
	UUID string `json:"uuid"`
}
type statsDocument struct {
	Stats map[string]map[string]float64 `json:"stats"`
}

// ReadProfiles combines the vanilla user cache, operator list, and per-player stats files.
func ReadProfiles(cachePath, opsPath, worldPath string) ([]Profile, error) {
	cache, err := readJSON[[]cacheEntry](cachePath)
	if err != nil {
		return nil, err
	}
	ops, err := readJSON[[]operatorEntry](opsPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	opIDs := map[string]bool{}
	for _, operator := range ops {
		opIDs[operator.UUID] = true
	}
	profiles := []Profile{}
	for _, entry := range cache {
		stats, err := readJSON[statsDocument](filepath.Join(worldPath, "players", "stats", entry.UUID+".json"))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		custom := stats.Stats["minecraft:custom"]
		ticks := custom["minecraft:play_time"]
		profiles = append(profiles, Profile{Name: entry.Name, Operator: opIDs[entry.UUID], PlayTime: fmt.Sprintf("%.1f h", ticks/72000), Deaths: number(custom["minecraft:deaths"]), Walked: fmt.Sprintf("%.1f km", custom["minecraft:walk_one_cm"]/100000), MobKills: number(custom["minecraft:mob_kills"]), playTicks: ticks})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].playTicks > profiles[j].playTicks })
	for index := range profiles {
		profiles[index].playTicks = 0
	}
	return profiles, nil
}
func readJSON[T any](path string) (T, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(data, &value)
	return value, err
}
func number(value float64) string { return fmt.Sprintf("%.0f", value) }
