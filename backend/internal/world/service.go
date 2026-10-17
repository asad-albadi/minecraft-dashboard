package world

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Metadata struct { Name string `json:"name"`; GameMode string `json:"gameMode"`; Difficulty string `json:"difficulty"`; Seed string `json:"seed"`; Dimension string `json:"dimension"`; Size string `json:"size"`; LastBackup string `json:"lastBackup"` }

func ReadMetadata(propertiesPath, worldPath string) (Metadata, error) {
	values, err := readProperties(propertiesPath); if err != nil { return Metadata{}, err }
	size, err := directorySize(worldPath); if err != nil { return Metadata{}, err }
	return Metadata{Name: values["level-name"], GameMode: title(values["gamemode"]), Difficulty: title(values["difficulty"]), Seed: masked(values["level-seed"]), Dimension:"Primary World: Overworld", Size: humanBytes(size), LastBackup:"Never"}, nil
}
func readProperties(path string) (map[string]string, error) { file, err := os.Open(path); if err != nil { return nil, err }; defer file.Close(); values := map[string]string{}; scanner := bufio.NewScanner(file); for scanner.Scan() { line := scanner.Text(); if strings.HasPrefix(line,"#") || !strings.Contains(line,"=") { continue }; pair := strings.SplitN(line,"=",2); values[pair[0]]=pair[1] }; return values, scanner.Err() }
func directorySize(path string) (int64, error) { var total int64; err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error { if err != nil { return err }; if !info.IsDir() { total += info.Size() }; return nil }); return total, err }
func humanBytes(bytes int64) string { if bytes < 1024 { return fmt.Sprintf("%d B", bytes) }; if bytes < 1024*1024 { return fmt.Sprintf("%.1f KB", float64(bytes)/1024) }; if bytes < 1024*1024*1024 { return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024)) }; return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024)) }
func title(value string) string { if value == "" { return "Unavailable" }; return strings.ToUpper(value[:1])+strings.ToLower(value[1:]) }
func masked(value string) string { if value == "" { return "Unavailable" }; return "••••••••" }
