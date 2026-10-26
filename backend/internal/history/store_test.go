package history

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPersistsBoundedEventsAndMonotonicIDs(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{}
	for i := 0; i < 503; i++ {
		lines = append(lines, fmt.Sprintf("[12:00:00] [Server thread/INFO]: <Alex> message %d", i))
	}
	if err := store.Append(lines); err != nil {
		t.Fatal(err)
	}
	events := store.Events()
	if len(events) != 500 || events[0].ID != 4 || events[499].ID != 503 {
		t.Fatalf("retention: %d %#v", len(events), events[0])
	}
	reopened, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reopened.Persistent() || len(reopened.Chat()) != 500 {
		t.Fatal("missing persisted chat")
	}
	if err := reopened.Append([]string{"[12:00:00] [Server thread/INFO]: Alex left the game"}); err != nil {
		t.Fatal(err)
	}
	events = reopened.Events()
	if events[499].ID != 504 || events[499].Kind != "leave" {
		t.Fatalf("restart identity: %#v", events[499])
	}
	events[0].Line = "changed"
	if reopened.Events()[0].Line == "changed" {
		t.Fatal("exposed mutable storage")
	}
}

func TestParsesOnlyServerEventsAndKeepsHonestTimestamps(t *testing.T) {
	store, err := NewStore("")
	if err != nil {
		t.Fatal(err)
	}
	if store.Persistent() {
		t.Fatal("memory store claims persistence")
	}
	err = store.Append([]string{
		"[12:00:00] [Server thread/INFO]: <Alex> Bee joined the game at 192.168.1.2",
		"[12:00:01] [Server thread/INFO]: Bee joined the game",
		"[12:00:02] [Server thread/INFO]: Bee was slain by Zombie",
		"[12:00:03] [Server thread/INFO]: Bee has made the advancement [Stone Age]",
		"[12:00:04] [Server thread/INFO]: Starting minecraft server version 1.21.1",
		"[12:00:05] [Server thread/INFO]: Stopping server",
		"[12:00:06] [Server thread/WARN]: Can't keep up!",
		"unstructured Bee joined the game",
		"[12:00:07] [RCON Listener/INFO]: secret 192.168.1.2",
	})
	if err != nil {
		t.Fatal(err)
	}
	events := store.Events()
	if len(events) != 7 {
		t.Fatalf("events: %#v", events)
	}
	for i, kind := range []string{"chat", "join", "death", "advancement", "startup", "shutdown", "warning"} {
		if events[i].Kind != kind {
			t.Fatalf("event %d: %#v", i, events[i])
		}
		if _, err := time.Parse(time.RFC3339Nano, events[i].ObservedAt); err != nil {
			t.Fatal(err)
		}
	}
	if events[0].Timestamp != "12:00:00" || strings.Contains(events[0].Line, "192.168.1.2") || events[0].Sender != "Alex" {
		t.Fatalf("unsafe/invented data: %#v", events[0])
	}
	if len(store.Chat()) != 1 {
		t.Fatal("nonchat classified as chat")
	}
}

func TestRejectsCorruptOrOversizedSnapshots(t *testing.T) {
	for _, body := range []string{`{"events":`, strings.Repeat(" ", 9*1024*1024)} {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "history.json"), []byte(body), 0600)
		if _, err := NewStore(dir); err == nil {
			t.Fatal("invalid snapshot accepted")
		}
	}
}
