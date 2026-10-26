package httpapi

import (
	"encoding/binary"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"io"
	"minecraft-dashboard/backend/internal/rcon"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func actionTestApp(t *testing.T) (*API, *fiber.App, chan string) {
	t.Helper()
	a, app := accountTestApp(t, "admin")
	a.Config.EnableChat = true
	a.Config.EnableConsole = true
	a.Config.DataPath = t.TempDir()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	commands := make(chan string, 10)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			func() {
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(time.Second))
				_, err := readActionPacket(conn)
				if err != nil {
					return
				}
				writeActionPacket(conn, 1, 2, "")
				command, err := readActionPacket(conn)
				if err != nil {
					return
				}
				audit, auditErr := os.ReadFile(filepath.Join(a.Config.DataPath, "audit.jsonl"))
				if auditErr != nil || !strings.Contains(string(audit), `"outcome":"attempt"`) {
					command = "missing-audit:" + command
				}
				commands <- command
				writeActionPacket(conn, 2, 0, "ok")
			}()
		}
	}()
	a.Config.RCONAddress = listener.Addr().String()
	a.Config.RCONPassword = "test-secret"
	a.RCON = rcon.New(a.Config.RCONAddress, a.Config.RCONPassword, time.Second)
	admin := func(c *fiber.Ctx) error { c.Locals("username", "alex"); c.Locals("role", "admin"); return c.Next() }
	app.Post("/chat", admin, a.sendChat)
	app.Post("/console", admin, a.runConsole)
	app.Get("/audit", admin, a.auditLog)
	return a, app, commands
}
func readActionPacket(r io.Reader) (string, error) {
	var prefix [4]byte
	if _, err := io.ReadFull(r, prefix[:]); err != nil {
		return "", err
	}
	raw := make([]byte, binary.LittleEndian.Uint32(prefix[:]))
	_, err := io.ReadFull(r, raw)
	if err != nil {
		return "", err
	}
	return string(raw[8 : len(raw)-2]), nil
}
func writeActionPacket(w io.Writer, id, typ uint32, body string) {
	raw := make([]byte, len(body)+14)
	binary.LittleEndian.PutUint32(raw, uint32(len(raw)-4))
	binary.LittleEndian.PutUint32(raw[4:], id)
	binary.LittleEndian.PutUint32(raw[8:], typ)
	copy(raw[12:], body)
	w.Write(raw)
}

func TestActionsAuditBeforeSendingAndRestrictConsole(t *testing.T) {
	a, app, commands := actionTestApp(t)
	for _, command := range []string{"op Alex", "stop", "seed", "list\nstop", "time set day"} {
		body, _ := json.Marshal(map[string]string{"command": command})
		res := authRequest(t, app, "POST", "/console", string(body), "https://dashboard.test", "application/json", nil)
		if res.StatusCode != 422 {
			t.Fatalf("unsafe command %q accepted: %d", command, res.StatusCode)
		}
	}
	res := authRequest(t, app, "POST", "/console", `{"command":"list"}`, "https://dashboard.test", "application/json", nil)
	if res.StatusCode != 200 {
		t.Fatalf("console=%d", res.StatusCode)
	}
	if got := <-commands; got != "list" {
		t.Fatal(got)
	}
	data, err := os.ReadFile(filepath.Join(a.Config.DataPath, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"actor":"alex"`) || !strings.Contains(string(data), `"outcome":"attempt"`) || !strings.Contains(string(data), `"outcome":"success"`) {
		t.Fatalf("audit=%s", data)
	}
}
func TestActionsFailClosedWithoutAuditAndWhenDisabled(t *testing.T) {
	a, app, commands := actionTestApp(t)
	a.Config.EnableChat = false
	if authRequest(t, app, "POST", "/chat", `{"message":"hello"}`, "https://dashboard.test", "application/json", nil).StatusCode != 503 {
		t.Fatal("disabled chat accepted")
	}
	a.Config.EnableChat = true
	a.Config.DataPath = ""
	if authRequest(t, app, "POST", "/chat", `{"message":"hello"}`, "https://dashboard.test", "application/json", nil).StatusCode != 503 {
		t.Fatal("unaudited chat accepted")
	}
	select {
	case command := <-commands:
		t.Fatalf("unexpected command %s", command)
	default:
	}
}
func TestChatEscapesMessageAndRateLimitsActor(t *testing.T) {
	_, app, commands := actionTestApp(t)
	res := authRequest(t, app, "POST", "/chat", `{"message":"hello\n/stop"}`, "https://dashboard.test", "application/json", nil)
	if res.StatusCode != 200 {
		t.Fatal(res.StatusCode)
	}
	command := <-commands
	if command != `tellraw @a {"text":"[Dashboard Admin] hello /stop"}` {
		t.Fatal(command)
	}
	for i := 0; i < 10; i++ {
		res = authRequest(t, app, "POST", "/chat", `{"message":"hello"}`, "https://dashboard.test", "application/json", nil)
	}
	if res.StatusCode != 429 {
		t.Fatalf("rate limit=%d", res.StatusCode)
	}
}

func TestAuditRejectsProtectedStorageAndSymlink(t *testing.T) {
	a, app, commands := actionTestApp(t)
	a.Config.WorldPath = a.Config.DataPath
	if authRequest(t, app, "POST", "/console", `{"command":"list"}`, "https://dashboard.test", "application/json", nil).StatusCode != 503 {
		t.Fatal("world audit directory accepted")
	}
	a.Config.WorldPath = ""
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(a.Config.DataPath, "audit.jsonl")); err != nil {
		t.Fatal(err)
	}
	if authRequest(t, app, "POST", "/console", `{"command":"list"}`, "https://dashboard.test", "application/json", nil).StatusCode != 503 {
		t.Fatal("audit symlink accepted")
	}
	data, _ := os.ReadFile(target)
	if string(data) != "preserve" {
		t.Fatal("symlink target changed")
	}
	select {
	case command := <-commands:
		t.Fatalf("unexpected command %s", command)
	default:
	}
}
func TestAuditRotationAndRead(t *testing.T) {
	a, app, _ := actionTestApp(t)
	entry := auditEntry{Time: time.Now(), Actor: "alex", Action: "console", Command: strings.Repeat("x", 900), Outcome: "attempt"}
	a.Actions.mu.Lock()
	for i := 0; i < 1100; i++ {
		if err := a.appendAudit(entry); err != nil {
			a.Actions.mu.Unlock()
			t.Fatal(err)
		}
	}
	a.Actions.mu.Unlock()
	for _, name := range []string{"audit.jsonl", "audit.jsonl.1"} {
		info, err := os.Stat(filepath.Join(a.Config.DataPath, name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() > auditMaxBytes {
			t.Fatal("audit exceeded size bound")
		}
	}
	response := authRequest(t, app, "GET", "/audit", "", "", "", nil)
	var entries []auditEntry
	if err := json.NewDecoder(response.Body).Decode(&entries); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || len(entries) != 200 {
		t.Fatalf("audit response=%d count=%d", response.StatusCode, len(entries))
	}
}
