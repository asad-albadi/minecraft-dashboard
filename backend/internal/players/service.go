// Package players reads vanilla player state from read-only Minecraft data.
package players

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

type Profile struct {
	Name       string `json:"name"`
	Operator   bool   `json:"operator"`
	PlayTime   string `json:"playTime"`
	Deaths     string `json:"deaths"`
	Walked     string `json:"walked"`
	MobKills   string `json:"mobKills"`
	JoinEvents string `json:"joinEvents,omitempty"`
	FirstJoin  string `json:"firstJoin,omitempty"`
	LastJoin   string `json:"lastJoin,omitempty"`
	playTicks  float64
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
func ReadProfiles(cachePath, opsPath, worldPath, logDirectory string) ([]Profile, error) {
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
	history, err := joinHistory(logDirectory)
	if err != nil && logDirectory != "" {
		return nil, err
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
		profile := Profile{Name: entry.Name, Operator: opIDs[entry.UUID], PlayTime: fmt.Sprintf("%.1f h", ticks/72000), Deaths: number(custom["minecraft:deaths"]), Walked: fmt.Sprintf("%.1f km", custom["minecraft:walk_one_cm"]/100000), MobKills: number(custom["minecraft:mob_kills"]), playTicks: ticks}
		if events := history[entry.Name]; events.count > 0 {
			profile.JoinEvents = number(float64(events.count))
			profile.FirstJoin = events.first
			profile.LastJoin = events.last
		}
		profiles = append(profiles, profile)
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

type joins struct {
	count int
	first string
	last  string
}

var archivedLogName = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-\d+\.log\.gz$`)
var joinLine = regexp.MustCompile(`(?:System chat: )?([A-Za-z0-9_]{1,16}) joined the game`)

func joinHistory(directory string) (map[string]joins, error) {
	history := map[string]joins{}
	if directory == "" {
		return history, nil
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return history, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.Name() == "latest.log" {
			info, statErr := entry.Info()
			if statErr != nil {
				return nil, statErr
			}
			file, openErr := os.Open(filepath.Join(directory, entry.Name()))
			if openErr != nil {
				return nil, openErr
			}
			scanErr := recordJoins(file, info.ModTime().Format("2006-01-02"), history)
			file.Close()
			if scanErr != nil {
				return nil, scanErr
			}
			continue
		}
		match := archivedLogName.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		file, err := os.Open(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		reader, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		date := match[1]
		scanErr := recordJoins(reader, date, history)
		closeErr := reader.Close()
		file.Close()
		if scanErr != nil {
			return nil, scanErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
	}
	return history, nil
}

func recordJoins(reader io.Reader, date string, history map[string]joins) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		joined := joinLine.FindStringSubmatch(scanner.Text())
		if joined == nil {
			continue
		}
		event := history[joined[1]]
		event.count++
		if event.first == "" || date < event.first {
			event.first = date
		}
		if date > event.last {
			event.last = date
		}
		history[joined[1]] = event
	}
	return scanner.Err()
}
