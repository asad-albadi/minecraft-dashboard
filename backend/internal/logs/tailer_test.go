package logs

import (
	"reflect"
	"testing"
)

func TestNewLinesReturnsOnlyAppendedLines(t *testing.T) {
	previous := []string{"one", "two"}
	current := []string{"one", "two", "three", "four"}

	got := NewLines(previous, current)
	want := []string{"three", "four"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NewLines() = %v, want %v", got, want)
	}
}

func TestNewLinesKeepsFreshLinesAfterLogRotation(t *testing.T) {
	previous := []string{"old one", "old two"}
	current := []string{"new one", "new two"}

	got := NewLines(previous, current)
	if !reflect.DeepEqual(got, current) {
		t.Fatalf("NewLines() = %v, want %v", got, current)
	}
}

func TestServerVersionReturnsTheLatestStartupVersion(t *testing.T) {
	lines := []string{
		"[10:00] [Server thread/INFO]: Starting minecraft server version 1.20.4",
		"[10:01] [Server thread/INFO]: Done",
		"[11:00] [Server thread/INFO]: Starting minecraft server version 1.21.5",
	}

	if got := ServerVersion(lines); got != "1.21.5" {
		t.Fatalf("ServerVersion() = %q, want %q", got, "1.21.5")
	}
}
