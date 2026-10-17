package world

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Metadata struct {
	Name       string `json:"name"`
	GameMode   string `json:"gameMode"`
	Difficulty string `json:"difficulty"`
	Seed       string `json:"seed"`
	Dimension  string `json:"dimension"`
	Size       string `json:"size"`
	LastBackup string `json:"lastBackup"`
	Day        string `json:"day,omitempty"`
	Time       string `json:"time,omitempty"`
	Weather    string `json:"weather,omitempty"`
}

func ReadMetadata(propertiesPath, worldPath string) (Metadata, error) {
	values, err := readProperties(propertiesPath)
	if err != nil {
		return Metadata{}, err
	}
	size, err := directorySize(worldPath)
	if err != nil {
		return Metadata{}, err
	}
	return Metadata{Name: values["level-name"], GameMode: title(values["gamemode"]), Difficulty: title(values["difficulty"]), Seed: masked(values["level-seed"]), Dimension: "Primary World: Overworld", Size: humanBytes(size), LastBackup: "Never"}, nil
}
func readProperties(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		pair := strings.SplitN(line, "=", 2)
		values[pair[0]] = pair[1]
	}
	return values, scanner.Err()
}
func directorySize(path string) (int64, error) {
	var total int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}
func humanBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}
func title(value string) string {
	if value == "" {
		return "Unavailable"
	}
	return strings.ToUpper(value[:1]) + strings.ToLower(value[1:])
}
func masked(value string) string {
	if value == "" {
		return "Unavailable"
	}
	return "••••••••"
}

var timePattern = regexp.MustCompile(`(?i)time is:?\s+(\d+)`)
var timelinePattern = regexp.MustCompile(`(?i)timeline\s+minecraft:day\s+is\s+at\s+(\d+)\s+tick`)

// ParseMinecraftTime converts the vanilla `time query daytime` response into
// Minecraft's 24-hour clock. Minecraft tick zero is 06:00, not midnight.
func ParseMinecraftTime(response string) (day, clock string, ok bool) {
	match := timePattern.FindStringSubmatch(response)
	timeline := false
	if match == nil {
		match = timelinePattern.FindStringSubmatch(response)
		timeline = true
	}
	if match == nil {
		return "", "", false
	}
	ticks, err := strconv.Atoi(match[1])
	if err != nil || ticks < 0 {
		return "", "", false
	}
	if !timeline {
		day = fmt.Sprintf("Day %d", ticks/24000)
	}
	clockTicks := (ticks + 6000) % 24000
	hours := clockTicks / 1000
	minutes := (clockTicks % 1000) * 60 / 1000
	return day, fmt.Sprintf("%02d:%02d", hours, minutes), true
}

// ParseWeather accepts only the vanilla weather query values, so an
// unrecognised server/plugin response remains unavailable instead of guessed.
func ParseWeather(response string) (string, bool) {
	response = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(response)), ".")
	switch response {
	case "the weather is clear", "the weather is: clear":
		return "Clear Skies", true
	case "the weather is raining", "the weather is: raining":
		return "Rain", true
	case "the weather is thundering", "the weather is: thundering":
		return "Thunderstorm", true
	default:
		return "", false
	}
}
