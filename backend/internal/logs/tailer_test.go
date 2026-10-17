package logs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTailReturnsLatestLinesAndMasksIPAddresses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "latest.log")
	if err := os.WriteFile(path, []byte("one\nRCON Client noise\nplayer joined from 192.168.1.9\nthree\n"), 0600); err != nil { t.Fatal(err) }
	lines, err := Tail(path, 2)
	if err != nil { t.Fatal(err) }
	if len(lines) != 2 || lines[0] != "player joined from <ip>" || lines[1] != "three" { t.Fatalf("unexpected lines: %#v", lines) }
}
