// Package backups inventories explicitly configured archives without modifying them.
package backups

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type Entry struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Type       string    `json:"type"`
	Integrity  string    `json:"integrity"`
}

// List refuses oversized directories rather than presenting an incomplete inventory.
func List(path string) ([]Entry, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	entries, e := f.ReadDir(1001)
	if e != nil && e != io.EOF {
		return nil, e
	}
	if len(entries) > 1000 {
		return nil, fmt.Errorf("backup directory exceeds 1000 entries")
	}
	out := []Entry{}
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			continue
		}
		kind := ""
		for _, suffix := range []string{".tar.gz", ".tar.zst", ".tgz", ".zip", ".tar"} {
			if strings.HasSuffix(strings.ToLower(entry.Name()), suffix) {
				kind = strings.TrimPrefix(suffix, ".")
				break
			}
		}
		if kind == "" {
			continue
		}
		info, e := entry.Info()
		if e != nil {
			return nil, e
		}
		if !info.Mode().IsRegular() {
			continue
		}
		out = append(out, Entry{entry.Name(), info.Size(), info.ModTime().UTC(), kind, "Unknown"})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ModifiedAt.After(out[j].ModifiedAt) })
	return out, nil
}
