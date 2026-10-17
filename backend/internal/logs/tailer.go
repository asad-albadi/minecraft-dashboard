package logs

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var ipAddress = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
var uuid = regexp.MustCompile(`\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
var coordinates = regexp.MustCompile(`\(-?\d+(?:\.\d+)?,\s*-?\d+(?:\.\d+)?,\s*-?\d+(?:\.\d+)?\)`)

func Tail(path string, limit int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]string, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "RCON Client") || strings.Contains(line, "RCON Listener") {
			continue
		}
		line = coordinates.ReplaceAllString(uuid.ReplaceAllString(ipAddress.ReplaceAllString(line, "<ip>"), "<uuid>"), "<coordinates>")
		lines = append(lines, line)
		if len(lines) > limit {
			lines = lines[1:]
		}
	}
	return lines, scanner.Err()
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
