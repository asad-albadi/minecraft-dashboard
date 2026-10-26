package httpapi

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"io"
	"log"
	"minecraft-dashboard/backend/internal/chat"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const auditMaxBytes = 1024 * 1024

// Authentication stays available when optional audit storage is absent or broken.
func (a *API) auditSession(actor, action string) {
	if a.Config.DataPath == "" {
		return
	}
	a.Actions.mu.Lock()
	defer a.Actions.mu.Unlock()
	if err := a.appendAudit(auditEntry{Time: time.Now().UTC(), Actor: actor, Action: action, Outcome: "success"}); err != nil {
		log.Printf("dashboard session audit unavailable")
	}
}

type actionManager struct {
	mu       sync.Mutex
	attempts map[string]loginWindow
}
type auditEntry struct {
	Time    time.Time `json:"time"`
	Actor   string    `json:"actor"`
	Action  string    `json:"action"`
	Command string    `json:"command,omitempty"`
	Outcome string    `json:"outcome"`
}

func (a *API) sendChat(c *fiber.Ctx) error {
	if !a.Config.EnableChat {
		return c.Status(503).JSON(fiber.Map{"error": "chat capability is not configured"})
	}
	var request struct {
		Message string `json:"message"`
	}
	if len(c.Body()) > 4096 || json.Unmarshal(c.Body(), &request) != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	command, err := chat.TellrawCommand(request.Message)
	if err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}
	return a.executeAction(c, "chat", command)
}
func (a *API) runConsole(c *fiber.Ctx) error {
	if !a.Config.EnableConsole {
		return c.Status(503).JSON(fiber.Map{"error": "console capability is not configured"})
	}
	var request struct {
		Command string `json:"command"`
	}
	if len(c.Body()) > 256 || json.Unmarshal(c.Body(), &request) != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	switch request.Command {
	case "list", "time query daytime", "time query day":
	default:
		return c.Status(422).JSON(fiber.Map{"error": "command must be list, time query daytime, or time query day"})
	}
	return a.executeAction(c, "console", request.Command)
}
func (a *API) executeAction(c *fiber.Ctx, action, command string) error {
	actor, ok := c.Locals("username").(string)
	if !ok || actor == "" || c.Locals("role") != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "administrator access required"})
	}
	if !a.csrf(c) {
		return c.Status(403).JSON(fiber.Map{"error": "request origin or content type rejected"})
	}
	if a.RCON == nil || a.Config.RCONAddress == "" || a.Config.RCONPassword == "" {
		return c.Status(503).JSON(fiber.Map{"error": "RCON is not configured"})
	}
	a.Actions.mu.Lock()
	defer a.Actions.mu.Unlock()
	now := time.Now()
	if a.Actions.attempts == nil {
		a.Actions.attempts = make(map[string]loginWindow)
	}
	for key, w := range a.Actions.attempts {
		if !now.Before(w.expires) {
			delete(a.Actions.attempts, key)
		}
	}
	w, exists := a.Actions.attempts[actor]
	if w.count >= 10 || (!exists && len(a.Actions.attempts) >= 100) {
		c.Set("Retry-After", "60")
		return c.Status(429).JSON(fiber.Map{"error": "too many actions; try again later"})
	}
	if !exists {
		w.expires = now.Add(time.Minute)
	}
	w.count++
	a.Actions.attempts[actor] = w
	entry := auditEntry{Time: now.UTC(), Actor: actor, Action: action, Command: command, Outcome: "attempt"}
	if err := a.appendAudit(entry); err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "audit storage unavailable; command was not sent"})
	}
	response, err := a.RCON.Command(c.Context(), command)
	entry.Time = time.Now().UTC()
	entry.Outcome = "success"
	if err != nil {
		entry.Outcome = "failed_or_unknown"
	}
	if auditErr := a.appendAudit(entry); auditErr != nil {
		return c.Status(502).JSON(fiber.Map{"error": "command may have run but its outcome could not be audited; do not retry automatically"})
	}
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "RCON delivery failed or outcome is unknown; do not retry automatically"})
	}
	return c.JSON(fiber.Map{"ok": true, "response": response})
}

func containedPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func (a *API) auditDirectory() (string, error) {
	dir := a.Config.DataPath
	if dir == "" || !filepath.IsAbs(dir) || filepath.Clean(dir) == "/" {
		return "", errors.New("configure dedicated dashboard storage")
	}
	dir = filepath.Clean(dir)
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	if resolved != dir {
		return "", errors.New("audit directory must not contain symlinks")
	}
	for _, source := range []string{a.Config.WorldPath, a.Config.LogDirectory, a.Config.LogPath, a.Config.PropertiesPath, a.Config.UserCachePath, a.Config.OpsPath, a.Config.AccountsFile, a.Config.BackupPath} {
		if source == "" {
			continue
		}
		source, _ = filepath.Abs(source)
		if real, err := filepath.EvalSymlinks(source); err == nil {
			source = real
		}
		if containedPath(source, dir) || containedPath(dir, source) {
			return "", errors.New("audit storage overlaps protected data")
		}
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return "", errors.New("audit directory unavailable")
	}
	return dir, nil
}
func openAudit(path string, flags int) (*os.File, error) {
	f, err := os.OpenFile(path, flags|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		f.Close()
		return nil, errors.New("audit file must be private regular file")
	}
	return f, nil
}

// appendAudit is called with Actions.mu held. Only this dashboard's two files rotate.
func (a *API) appendAudit(entry auditEntry) error {
	dir, err := a.auditDirectory()
	if err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := filepath.Join(dir, "audit.jsonl")
	f, err := openAudit(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	if info.Size()+int64(len(data)) > auditMaxBytes {
		if err := f.Close(); err != nil {
			return err
		}
		rotated := path + ".1"
		if stat, err := os.Lstat(rotated); err == nil && (!stat.Mode().IsRegular() || stat.Mode().Perm()&0077 != 0) {
			return errors.New("invalid rotated audit file")
		}
		if err := os.Rename(path, rotated); err != nil {
			return err
		}
		f, err = openAudit(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY)
		if err != nil {
			return err
		}
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}
func (a *API) auditLog(c *fiber.Ctx) error {
	if c.Locals("role") != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "administrator access required"})
	}
	a.Actions.mu.Lock()
	defer a.Actions.mu.Unlock()
	dir, err := a.auditDirectory()
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "audit storage unavailable"})
	}
	entries := make([]auditEntry, 0)
	for _, name := range []string{"audit.jsonl.1", "audit.jsonl"} {
		f, err := openAudit(filepath.Join(dir, name), os.O_RDONLY)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return c.Status(503).JSON(fiber.Map{"error": "audit storage unavailable"})
		}
		scanner := bufio.NewScanner(io.LimitReader(f, auditMaxBytes+1))
		for scanner.Scan() {
			var entry auditEntry
			if json.Unmarshal(scanner.Bytes(), &entry) != nil {
				f.Close()
				return c.Status(503).JSON(fiber.Map{"error": "audit data unreadable"})
			}
			entries = append(entries, entry)
			if len(entries) > 200 {
				entries = entries[1:]
			}
		}
		err = scanner.Err()
		f.Close()
		if err != nil {
			return c.Status(503).JSON(fiber.Map{"error": "audit data unreadable"})
		}
	}
	return c.JSON(entries)
}
