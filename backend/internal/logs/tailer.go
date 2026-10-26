package logs

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

var ipAddress = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
var ipv6Address = regexp.MustCompile(`(?i)[0-9a-f:.]*:[0-9a-f:.]+(?:%[a-z0-9_-]+)?`)
var uuid = regexp.MustCompile(`\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
var coordinates = regexp.MustCompile(`\(-?\d+(?:\.\d+)?,\s*-?\d+(?:\.\d+)?,\s*-?\d+(?:\.\d+)?\)`)
var serverVersion = regexp.MustCompile(`Starting minecraft server version ([^\s]+)`)

func Tail(path string, limit int) ([]string, error) {
	r := NewReader(path, limit)
	defer r.Close()
	return r.Next()
}

// Sanitize removes private endpoints and identifiers while retaining timestamps.
func Sanitize(line string) string {
	if strings.Contains(line, "RCON Client") || strings.Contains(line, "RCON Listener") {
		return ""
	}
	line = ipv6Address.ReplaceAllStringFunc(line, func(s string) string {
		address := strings.SplitN(s, "%", 2)[0]
		if net.ParseIP(address) != nil {
			return "<ip>"
		}
		return s
	})
	line = ipAddress.ReplaceAllStringFunc(line, func(s string) string {
		if net.ParseIP(s) != nil {
			return "<ip>"
		}
		return s
	})
	return coordinates.ReplaceAllString(uuid.ReplaceAllString(line, "<uuid>"), "<coordinates>")
}

type Chat struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

var chatPattern = regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO\]: (?:\[Not Secure\] )?<([A-Za-z0-9_]{1,16})> (.*)$`)

func ParseChat(line string) (Chat, bool) {
	m := chatPattern.FindStringSubmatch(Sanitize(line))
	if m == nil {
		return Chat{}, false
	}
	return Chat{m[2], m[3], m[1]}, true
}

const readBound = 256 * 1024
const lineBound = 16 * 1024

// Reader keeps an offset rather than comparing text: repeated messages are events.
// Next reads at most 256 KiB and returns complete lines. Call again to drain bursts.
type Reader struct {
	path     string
	limit    int
	file     *os.File
	offset   int64
	pending  []byte
	dropping bool
	anchor   []byte
}

func NewReader(path string, limit int) *Reader {
	if limit < 1 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	return &Reader{path: path, limit: limit}
}
func (r *Reader) Close() error {
	if r.file != nil {
		err := r.file.Close()
		r.file = nil
		return err
	}
	return nil
}
func (r *Reader) Next() ([]string, error) {
	info, err := os.Stat(r.path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("log source is not a regular file")
	}
	initial := r.file == nil
	if r.file != nil {
		old, e := r.file.Stat()
		if e != nil {
			return nil, e
		}
		if !os.SameFile(old, info) {
			// Drain the old inode first so append-before-rename events are not lost.
			if old.Size() > r.offset {
				return r.read(false)
			}
			r.Close()
			r.offset = 0
			r.pending = nil
			r.dropping = false
			r.anchor = nil
		} else if info.Size() < r.offset || r.rewritten() {
			r.offset = 0
			r.pending = nil
			r.dropping = false
			r.anchor = nil
			r.file.Seek(0, io.SeekStart)
		}
	}
	if r.file == nil {
		r.file, err = os.Open(r.path)
		if err != nil {
			return nil, err
		}
		if initial && info.Size() > readBound {
			r.offset = info.Size() - readBound
			r.dropping = true
		}
		if _, err = r.file.Seek(r.offset, io.SeekStart); err != nil {
			return nil, err
		}
	}
	return r.read(initial)
}
func (r *Reader) read(initial bool) ([]string, error) {
	buf := make([]byte, readBound)
	n, err := r.file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	r.offset += int64(n)
	if n > 0 {
		r.anchor = append(r.anchor, buf[:n]...)
		if len(r.anchor) > 64 {
			r.anchor = append([]byte(nil), r.anchor[len(r.anchor)-64:]...)
		}
	}
	lines := []string{}
	for _, b := range buf[:n] {
		if b == '\n' {
			if !r.dropping {
				line := Sanitize(string(bytes.TrimSuffix(r.pending, []byte{'\r'})))
				if line != "" {
					lines = append(lines, line)
				}
			}
			r.pending = nil
			r.dropping = false
			continue
		}
		if r.dropping {
			continue
		}
		if len(r.pending) >= lineBound {
			r.pending = nil
			r.dropping = true
			continue
		}
		r.pending = append(r.pending, b)
	}
	if initial && len(lines) > r.limit {
		lines = lines[len(lines)-r.limit:]
	}
	return lines, nil
}

// A small preceding-byte checkpoint also catches copy-truncate followed by
// regrowth beyond the previous offset between polling intervals.
func (r *Reader) rewritten() bool {
	if len(r.anchor) == 0 {
		return false
	}
	b := make([]byte, len(r.anchor))
	n, e := r.file.ReadAt(b, r.offset-int64(len(b)))
	return e != nil || n != len(b) || !bytes.Equal(b, r.anchor)
}

// NewLines returns the lines added since the prior sanitized tail. If a log was
// rotated, there is no overlap and the fresh tail is emitted intact.
func NewLines(previous, current []string) []string {
	maxOverlap := len(previous)
	if len(current) < maxOverlap {
		maxOverlap = len(current)
	}
	for overlap := maxOverlap; overlap > 0; overlap-- {
		matches := true
		for index := 0; index < overlap; index++ {
			if previous[len(previous)-overlap+index] != current[index] {
				matches = false
				break
			}
		}
		if matches {
			return current[overlap:]
		}
	}
	return current
}

// ServerVersion returns the most recent version advertised by a vanilla server
// startup line, or an empty string when no such line is present.
func ServerVersion(lines []string) string {
	version := ""
	for _, line := range lines {
		if match := serverVersion.FindStringSubmatch(line); len(match) == 2 {
			version = match[1]
		}
	}
	return version
}
