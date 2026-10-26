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
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type Profile struct {
	UUID               string                        `json:"uuid"`
	Stats              map[string]map[string]float64 `json:"stats"`
	StatsStatus        string                        `json:"statsStatus"`
	Advancements       []Advancement                 `json:"advancements"`
	AdvancementsStatus string                        `json:"advancementsStatus"`
	Inventory          string                        `json:"inventory"`
	HistoryNote        string                        `json:"historyNote"`
	Name               string                        `json:"name"`
	Operator           bool                          `json:"operator"`
	PlayTime           string                        `json:"playTime"`
	Deaths             string                        `json:"deaths"`
	Walked             string                        `json:"walked"`
	MobKills           string                        `json:"mobKills"`
	JoinEvents         string                        `json:"joinEvents,omitempty"`
	FirstJoin          string                        `json:"firstJoin,omitempty"`
	LastJoin           string                        `json:"lastJoin,omitempty"`
	playTicks          float64
}
type Advancement struct {
	ID       string            `json:"id"`
	Category string            `json:"category"`
	Done     bool              `json:"done"`
	Date     string            `json:"date"`
	Progress string            `json:"progress"`
	Criteria map[string]string `json:"criteria"`
}

var validUUID = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

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
	history, historyErr := cachedJoinHistory(logDirectory)
	profiles := []Profile{}
	for _, entry := range cache {
		profile := Profile{UUID: entry.UUID, Name: entry.Name, Operator: opIDs[entry.UUID], PlayTime: "Unknown", Deaths: "Unknown", Walked: "Unknown", MobKills: "Unknown", JoinEvents: "Unknown", FirstJoin: "Unknown", LastJoin: "Unknown", Stats: map[string]map[string]float64{}, StatsStatus: "Unavailable", Advancements: []Advancement{}, AdvancementsStatus: "Unavailable", Inventory: "Unavailable", HistoryNote: "Observed joins in bounded retained logs only; not lifetime history. Dates use archive filenames or latest.log modification date; server timezone is unknown."}
		if validUUID.MatchString(entry.UUID) {
			if stats, err := readPlayerJSON[statsDocument](worldPath, "stats", entry.UUID); err == nil && stats.Stats != nil {
				profile.Stats = stats.Stats
				profile.StatsStatus = "Available"
				custom := stats.Stats["minecraft:custom"]
				if value, ok := custom["minecraft:play_time"]; ok {
					profile.PlayTime = fmt.Sprintf("%.1f h", value/72000)
					profile.playTicks = value
				}
				if value, ok := custom["minecraft:deaths"]; ok {
					profile.Deaths = number(value)
				}
				if value, ok := custom["minecraft:walk_one_cm"]; ok {
					profile.Walked = fmt.Sprintf("%.1f km", value/100000)
				}
				if value, ok := custom["minecraft:mob_kills"]; ok {
					profile.MobKills = number(value)
				}
			}
			if raw, err := readPlayerJSON[map[string]json.RawMessage](worldPath, "advancements", entry.UUID); err == nil && raw != nil {
				profile.AdvancementsStatus = "Available"
				for id, body := range raw {
					if !strings.Contains(id, ":") {
						continue
					}
					var record struct {
						Done     *bool             `json:"done"`
						Criteria map[string]string `json:"criteria"`
					}
					if json.Unmarshal(body, &record) != nil || record.Done == nil {
						continue
					}
					category := strings.SplitN(strings.SplitN(id, ":", 2)[1], "/", 2)[0]
					a := Advancement{ID: id, Category: category, Done: *record.Done, Date: "Unknown", Progress: "Unknown", Criteria: record.Criteria}
					if a.Done {
						a.Progress = "Complete"
						for _, date := range a.Criteria {
							if _, err := time.Parse("2006-01-02 15:04:05 -0700", date); err == nil && (a.Date == "Unknown" || date > a.Date) {
								a.Date = date
							}
						}
					}
					profile.Advancements = append(profile.Advancements, a)
				}
				sort.Slice(profile.Advancements, func(i, j int) bool { return profile.Advancements[i].ID < profile.Advancements[j].ID })
			}
		}
		if historyErr != nil {
			profile.HistoryNote += " History is incomplete: some retained logs were unreadable or exceeded scan limits. Counts and dates describe only observed joins."
		}
		if events := history[entry.Name]; events.count > 0 {
			profile.JoinEvents = number(float64(events.count))
			profile.FirstJoin = events.first
			profile.LastJoin = events.last
		}
		profiles = append(profiles, profile)
	}
	sort.SliceStable(profiles, func(i, j int) bool { return profiles[i].playTicks > profiles[j].playTicks })
	for index := range profiles {
		profiles[index].playTicks = 0
	}
	return profiles, nil
}
func readJSON[T any](path string) (T, error) {
	var value T
	file, err := os.Open(path)
	if err != nil {
		return value, err
	}
	defer file.Close()
	return decodeJSON[T](file)
}
func decodeJSON[T any](file *os.File) (T, error) {
	var value T
	info, err := file.Stat()
	if err != nil {
		return value, err
	}
	if !info.Mode().IsRegular() {
		return value, fmt.Errorf("not a regular JSON file")
	}
	data, err := io.ReadAll(io.LimitReader(file, 4*1024*1024+1))
	if err != nil {
		return value, err
	}
	if len(data) > 4*1024*1024 {
		return value, fmt.Errorf("JSON exceeds size limit")
	}
	err = json.Unmarshal(data, &value)
	return value, err
}

// Open descendants relative to directory descriptors. O_NOFOLLOW on every
// component prevents symlink escapes even if paths change during a read.
// The explicitly configured root is trusted and may itself be a mount/symlink.
func openSource(root string, parts ...string) (*os.File, error) {
	fd, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	for i, part := range parts {
		if part == "" || part == "." || part == ".." || strings.Contains(part, "/") {
			unix.Close(fd)
			return nil, fmt.Errorf("invalid source path")
		}
		flags := unix.O_RDONLY | unix.O_NOFOLLOW | unix.O_CLOEXEC | unix.O_NONBLOCK
		if i < len(parts)-1 {
			flags |= unix.O_DIRECTORY
		}
		next, err := unix.Openat(fd, part, flags, 0)
		unix.Close(fd)
		if err != nil {
			return nil, err
		}
		fd = next
	}
	file := os.NewFile(uintptr(fd), filepath.Join(append([]string{root}, parts...)...))
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, fmt.Errorf("not a regular source file")
	}
	return file, nil
}
func readPlayerJSON[T any](root, kind, id string) (T, error) {
	var zero T
	if !validUUID.MatchString(id) {
		return zero, fmt.Errorf("invalid UUID")
	}
	for _, parts := range [][]string{{kind, id + ".json"}, {"players", kind, id + ".json"}} {
		file, err := openSource(root, parts...)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return zero, err
		}
		value, err := decodeJSON[T](file)
		file.Close()
		return value, err
	}
	return zero, os.ErrNotExist
}
func number(value float64) string { return fmt.Sprintf("%.0f", value) }

type joins struct {
	count int
	first string
	last  string
}

var archivedLogName = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-\d+\.log\.gz$`)
var joinLine = regexp.MustCompile(`^\[\d{2}:\d{2}:\d{2}\] \[Server thread/INFO\]: (?:System chat: )?([A-Za-z0-9_]{1,16}) joined the game\s*$`)

var historyCache struct {
	sync.Mutex
	directory string
	until     time.Time
	history   map[string]joins
	err       error
}

func cachedJoinHistory(directory string) (map[string]joins, error) {
	historyCache.Lock()
	defer historyCache.Unlock()
	if historyCache.directory == directory && time.Now().Before(historyCache.until) {
		return historyCache.history, historyCache.err
	}
	history, err := joinHistory(directory)
	historyCache.directory = directory
	historyCache.until = time.Now().Add(30 * time.Second)
	historyCache.history = history
	historyCache.err = err
	return history, err
}

func joinHistory(directory string) (map[string]joins, error) {
	history := map[string]joins{}
	if directory == "" {
		return history, nil
	}
	dir, err := os.Open(directory)
	if err != nil {
		return history, err
	}
	defer dir.Close()
	entries, err := dir.ReadDir(1025)
	var partialErr error
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		if os.IsNotExist(err) {
			return history, nil
		}
		partialErr = err
	}
	if len(entries) > 1024 {
		entries = entries[:1024]
		partialErr = fmt.Errorf("log directory exceeds bounded history limit")
	}
	// Deterministic order; a bad archive must never discard or prevent other
	// independently readable observations.
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var totalBytes int64
	remainingBytes := int64(32 * 1024 * 1024)
	recordBounded := func(reader io.Reader, date string) error {
		limited := &io.LimitedReader{R: reader, N: remainingBytes + 1}
		err := recordJoins(limited, date, history)
		remainingBytes = limited.N - 1
		if remainingBytes < 0 {
			return fmt.Errorf("total decompressed history exceeds size limit")
		}
		return err
	}
	for _, entry := range entries {
		if remainingBytes <= 0 {
			partialErr = fmt.Errorf("total decompressed history exceeds size limit")
			break
		}
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			continue
		}
		if entry.Name() != "latest.log" && !archivedLogName.MatchString(entry.Name()) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			partialErr = infoErr
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if totalBytes+info.Size() > 32*1024*1024 {
			partialErr = fmt.Errorf("retained logs exceed history size limit")
			continue
		}
		totalBytes += info.Size()
		if entry.Name() == "latest.log" {
			info, statErr := entry.Info()
			if statErr != nil {
				partialErr = statErr
				continue
			}
			file, openErr := openSource(directory, entry.Name())
			if openErr != nil {
				partialErr = openErr
				continue
			}
			scanErr := recordBounded(file, info.ModTime().Format("2006-01-02"))
			file.Close()
			if scanErr != nil {
				partialErr = scanErr
			}
			continue
		}
		match := archivedLogName.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		file, err := openSource(directory, entry.Name())
		if err != nil {
			partialErr = err
			continue
		}
		reader, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			partialErr = err
			continue
		}
		date := match[1]
		scanErr := recordBounded(reader, date)
		closeErr := reader.Close()
		file.Close()
		if scanErr != nil {
			partialErr = scanErr
		}
		if closeErr != nil {
			partialErr = closeErr
		}
	}
	return history, partialErr
}

func recordJoins(reader io.Reader, date string, history map[string]joins) error {
	bounded := &io.LimitedReader{R: reader, N: 4*1024*1024 + 1}
	scanner := bufio.NewScanner(bounded)
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
	if bounded.N == 0 {
		return fmt.Errorf("log exceeds decompressed size limit")
	}
	return scanner.Err()
}
