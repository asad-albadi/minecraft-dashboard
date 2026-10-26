package logs

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestPrivacyPreservesClockAndChat(t *testing.T) {
	got := Sanitize("[12:34:56] [Server thread/INFO]: <Alex> hi 127.0.0.1 ::1 2001:db8::1")
	if got != "[12:34:56] [Server thread/INFO]: <Alex> hi <ip> <ip> <ip>" {
		t.Fatal(got)
	}
	chat, ok := ParseChat("[12:34:56] [Server thread/INFO]: [Not Secure] <Alex> hello")
	if !ok || chat.Sender != "Alex" || chat.Message != "hello" || chat.Timestamp != "12:34:56" {
		t.Fatalf("%+v %v", chat, ok)
	}
	if _, ok := ParseChat("[12:34:56] [Server thread/INFO]: Alex joined the game"); ok {
		t.Fatal("activity treated as chat")
	}
}

func TestReaderAppendRotationTruncateAndPartial(t *testing.T) {
	p := filepath.Join(t.TempDir(), "latest.log")
	os.WriteFile(p, []byte("same\nsame\n"), 0600)
	r := NewReader(p, 10)
	defer r.Close()
	got, err := r.Next()
	if err != nil || len(got) != 2 {
		t.Fatalf("%v %v", got, err)
	}
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0600)
	f.WriteString("same\npar")
	f.Close()
	got, err = r.Next()
	if err != nil || !reflect.DeepEqual(got, []string{"same"}) {
		t.Fatalf("%v %v", got, err)
	}
	f, _ = os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0600)
	f.WriteString("tial\n")
	f.Close()
	got, _ = r.Next()
	if !reflect.DeepEqual(got, []string{"partial"}) {
		t.Fatal(got)
	}
	os.Rename(p, p+".1")
	os.WriteFile(p, []byte("rotated\n"), 0600)
	got, _ = r.Next()
	if !reflect.DeepEqual(got, []string{"rotated"}) {
		t.Fatal(got)
	}
	os.WriteFile(p, []byte("new\n"), 0600)
	got, _ = r.Next()
	if !reflect.DeepEqual(got, []string{"new"}) {
		t.Fatal(got)
	}
}

func TestTailBoundsOversizedLines(t *testing.T) {
	p := filepath.Join(t.TempDir(), "latest.log")
	os.WriteFile(p, []byte(strings.Repeat("x", 2*1024*1024)+"\nlast\n"), 0600)
	got, err := Tail(p, 2)
	if err != nil || len(got) != 1 || got[0] != "last" {
		t.Fatalf("%v %v", got, err)
	}
}

func TestReaderDetectsTruncateAndRegrowBetweenPolls(t *testing.T) {
	p := filepath.Join(t.TempDir(), "latest.log")
	os.WriteFile(p, []byte("before\n"), 0600)
	r := NewReader(p, 10)
	defer r.Close()
	r.Next()
	os.WriteFile(p, []byte("after!\nextra\n"), 0600)
	got, err := r.Next()
	if err != nil || !reflect.DeepEqual(got, []string{"after!", "extra"}) {
		t.Fatalf("%v %v", got, err)
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

func TestTailRedactsIPv6Addresses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "latest.log")
	line := "[12:00] [Server thread/INFO]: Alex[/[2001:db8:85a3::8a2e:370:7334]:25565] logged in"
	if err := os.WriteFile(path, []byte(line+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	lines, err := Tail(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0] != "[12:00] [Server thread/INFO]: Alex[/[<ip>]:25565] logged in" {
		t.Fatalf("Tail() = %#v, expected IPv6 address to be redacted", lines)
	}
}
