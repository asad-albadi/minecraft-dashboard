package httpapi

import (
	"context"
	"minecraft-dashboard/backend/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestUnconfiguredMonitoringDoesNotInventOfflineOrPlayerCounts(t *testing.T) {
	a := New(config.Config{})
	defer a.Close()
	status := a.statusPayload(context.Background())
	if status["state"] != "unknown" {
		t.Fatalf("status = %#v", status)
	}
	if _, ok := status["players"]; ok {
		t.Fatal("invented player count")
	}
}

func TestHistoryDoesNotWriteInsideWorld(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "dashboard-history")
	a := New(config.Config{WorldPath: dir, DataPath: data})
	defer a.Close()
	if a.History.Persistent() {
		t.Fatal("history may write inside world")
	}
	if _, err := os.Stat(data); !os.IsNotExist(err) {
		t.Fatal("created protected directory")
	}
}
func TestConsoleSnapshotsAreOrderedAndCopies(t *testing.T) {
	a := New(config.Config{})
	defer a.Close()
	a.recordLines([]string{"one"})
	first := a.consoleSnapshot()
	a.recordLines([]string{"two"})
	second := a.consoleSnapshot()
	if first.Sequence >= second.Sequence || len(first.Lines) != 1 || len(second.Lines) != 2 {
		t.Fatal(first, second)
	}
}

func TestSettingsHaveGenericNameAndNoSecrets(t *testing.T) {
	a := New(config.Config{SessionSecret: "private", RCONPassword: "private-rcon"})
	defer a.Close()
	s := a.settingsPayload()
	if s["serverName"] != "Minecraft Server" {
		t.Fatal(s)
	}
	if _, ok := s["sessionSecret"]; ok {
		t.Fatal("secret exposed")
	}
}
