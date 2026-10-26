// Package history retains bounded, sanitized observations derived from server logs.
// It never reads or modifies Minecraft data. Timestamps lack a date/timezone unless
// present in the source; ObservedAt records dashboard receipt, not event occurrence.
package history

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/sys/unix"
	"minecraft-dashboard/backend/internal/logs"
)

const retained = 500
const maxSnapshot = 8 * 1024 * 1024
const maxLine = 2048

type Event struct {
	ID         uint64 `json:"id"`
	Kind       string `json:"kind"`
	Sender     string `json:"sender,omitempty"`
	Message    string `json:"message"`
	Line       string `json:"line"`
	Timestamp  string `json:"timestamp"`
	ObservedAt string `json:"observedAt"`
}
type snapshot struct {
	NextID uint64  `json:"nextId"`
	Events []Event `json:"events"`
	Chat   []Event `json:"chat"`
}
type Store struct {
	mu        sync.RWMutex
	directory string
	state     snapshot
}

func NewStore(directory string) (*Store, error) {
	store := &Store{directory: directory, state: snapshot{NextID: 1, Events: []Event{}, Chat: []Event{}}}
	if directory == "" {
		return store, nil
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, "history.json")
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if os.IsNotExist(err) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("history snapshot is not a regular file")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxSnapshot+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxSnapshot {
		return nil, fmt.Errorf("history snapshot exceeds limit")
	}
	if err = json.Unmarshal(data, &store.state); err != nil {
		return nil, fmt.Errorf("invalid history snapshot: %w", err)
	}
	if err = validate(store.state); err != nil {
		return nil, err
	}
	return store, nil
}
func validate(state snapshot) error {
	if state.NextID == 0 || len(state.Events) > retained || len(state.Chat) > retained {
		return fmt.Errorf("invalid history snapshot bounds")
	}
	for _, list := range [][]Event{state.Events, state.Chat} {
		var previous uint64
		for _, event := range list {
			if event.ID <= previous || event.ID >= state.NextID || len(event.Line) > maxLine || len(event.Message) > maxLine {
				return fmt.Errorf("invalid history event bounds")
			}
			if _, err := time.Parse(time.RFC3339Nano, event.ObservedAt); err != nil {
				return fmt.Errorf("invalid observation time")
			}
			parsed, ok := parse(event.Line)
			if !ok || parsed.Kind != event.Kind || parsed.Sender != event.Sender || parsed.Message != event.Message || parsed.Timestamp != event.Timestamp || parsed.Line != event.Line {
				return fmt.Errorf("invalid history event source")
			}
			previous = event.ID
		}
	}
	for _, event := range state.Chat {
		if event.Kind != "chat" {
			return fmt.Errorf("invalid chat history")
		}
	}
	return nil
}

func (s *Store) Persistent() bool { return s.directory != "" }
func (s *Store) Events() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Event{}, s.state.Events...)
}
func (s *Store) Chat() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Event{}, s.state.Chat...)
}

// Reconcile recovers the bounded startup tail after the latest persisted event.
// Source timestamps have no date: identical lines across rotations are ambiguous.
func (s *Store) Reconcile(lines []string) error {
	previous := s.Events()
	if len(previous) > 0 {
		last := previous[len(previous)-1].Line
		for i := len(lines) - 1; i >= 0; i-- {
			if lines[i] == last {
				return s.Append(lines[i+1:])
			}
		}
	}
	return s.Append(lines)
}

// Append expects only new lines, in source order. Repeated identical messages are
// distinct events; the caller owns log offsets and restart replay prevention.
func (s *Store) Append(lines []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := snapshot{NextID: s.state.NextID, Events: append([]Event{}, s.state.Events...), Chat: append([]Event{}, s.state.Chat...)}
	for _, line := range lines {
		event, ok := parse(line)
		if !ok {
			continue
		}
		if next.NextID == ^uint64(0) {
			return fmt.Errorf("history identifiers exhausted")
		}
		event.ID = next.NextID
		next.NextID++
		event.ObservedAt = time.Now().UTC().Format(time.RFC3339Nano)
		next.Events = keep(append(next.Events, event))
		if event.Kind == "chat" {
			next.Chat = keep(append(next.Chat, event))
		}
	}
	if next.NextID == s.state.NextID {
		return nil
	}
	if s.Persistent() {
		if err := s.save(next); err != nil {
			return err
		}
	}
	s.state = next
	return nil
}
func keep(events []Event) []Event {
	if len(events) > retained {
		return events[len(events)-retained:]
	}
	return events
}

func (s *Store) save(state snapshot) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if len(data) > maxSnapshot {
		return fmt.Errorf("history snapshot exceeds limit")
	}
	file, err := os.CreateTemp(s.directory, ".history-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(s.directory, "history.json"))
}

var envelope = regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2})\] \[([^\]]+)/(INFO|WARN|ERROR)\]: (.*)$`)
var playerEvent = regexp.MustCompile(`^([A-Za-z0-9_]{1,16}) (.*)$`)
var advancement = regexp.MustCompile(`^has (?:made the advancement|completed the challenge|reached the goal) \[.+\]$`)
var death = regexp.MustCompile(`^(?:was (?:slain|shot|fireballed|killed|pummeled|blown up|squashed|impaled|speared|roasted|obliterated)|drowned|burned to death|went up in flames|tried to swim in lava|fell (?:from|off|out)|hit the ground too hard|blew up|suffocated in a wall|starved to death|withered away|froze to death|experienced kinetic energy|was doomed to fall|walked into danger|discovered the floor was lava|didn't want to live in the same world|left the confines of this world|was poked to death|died)(?:\b|$)`)

func parse(line string) (Event, bool) {
	if len(line) > maxLine {
		return Event{}, false
	}
	line = logs.Sanitize(strings.TrimSuffix(line, "\r"))
	for _, r := range line {
		if unicode.IsControl(r) {
			return Event{}, false
		}
	}
	match := envelope.FindStringSubmatch(line)
	if match == nil {
		return Event{}, false
	}
	if _, err := time.Parse("15:04:05", match[1]); err != nil {
		return Event{}, false
	}
	event := Event{Timestamp: match[1], Line: line, Message: match[4]}
	if parsed, ok := logs.ParseChat(line); ok {
		event.Kind = "chat"
		event.Sender = parsed.Sender
		event.Message = parsed.Message
		return event, true
	}
	if match[3] == "WARN" || match[3] == "ERROR" {
		event.Kind = "warning"
		return event, true
	}
	if match[2] != "Server thread" {
		return Event{}, false
	}
	body := match[4]
	if strings.HasPrefix(body, "Starting minecraft server version ") || strings.HasPrefix(body, "Done (") {
		event.Kind = "startup"
		return event, true
	}
	if body == "Stopping server" {
		event.Kind = "shutdown"
		return event, true
	}
	parts := playerEvent.FindStringSubmatch(strings.TrimPrefix(body, "System chat: "))
	if parts == nil {
		return Event{}, false
	}
	event.Sender = parts[1]
	switch {
	case parts[2] == "joined the game":
		event.Kind = "join"
	case parts[2] == "left the game":
		event.Kind = "leave"
	case advancement.MatchString(parts[2]):
		event.Kind = "advancement"
	case death.MatchString(parts[2]):
		event.Kind = "death"
	default:
		return Event{}, false
	}
	return event, true
}
