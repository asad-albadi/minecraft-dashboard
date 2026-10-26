package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"minecraft-dashboard/backend/internal/config"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func accountTestApp(t *testing.T, role string) (*API, *fiber.App) {
	t.Helper()
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "accounts.json")
	data, _ := json.Marshal([]account{{Username: "alex", PasswordHash: hash, Role: role}})
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	a := &API{Config: config.Config{SessionSecret: strings.Repeat("s", 32), AccountsFile: path, CookieSecure: true, PublicOrigin: "https://dashboard.test"}}
	app := fiber.New()
	a.RegisterSessionRoutes(app)
	app.Post("/admin", a.requireSession, a.requireAdmin, func(c *fiber.Ctx) error { return c.SendStatus(204) })
	return a, app
}
func authRequest(t *testing.T, app *fiber.App, method, path, body, origin, content string, cookie *http.Cookie) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, "https://dashboard.test"+path, strings.NewReader(body))
	req.Header.Set("Origin", origin)
	req.Header.Set("Content-Type", content)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	res, err := app.Test(req, 10000)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}
func TestNamedSessionIdentityLogoutAndExpiry(t *testing.T) {
	a, app := accountTestApp(t, "admin")
	login := authRequest(t, app, "POST", "/api/v1/session", `{"username":"alex","password":"correct-password"}`, "https://dashboard.test", "application/json", nil)
	if login.StatusCode != 204 {
		t.Fatalf("login=%d", login.StatusCode)
	}
	cookie := login.Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("unsafe cookie")
	}
	get := authRequest(t, app, "GET", "/api/v1/session", "", "", "", cookie)
	var who Identity
	json.NewDecoder(get.Body).Decode(&who)
	if get.StatusCode != 200 || who.Username != "alex" || who.Role != "admin" {
		t.Fatalf("identity=%+v status=%d", who, get.StatusCode)
	}
	admin := authRequest(t, app, "POST", "/admin", "{}", "https://dashboard.test", "application/json", cookie)
	if admin.StatusCode != 204 {
		t.Fatalf("admin=%d", admin.StatusCode)
	}
	logout := authRequest(t, app, "DELETE", "/api/v1/session", "{}", "https://dashboard.test", "application/json", cookie)
	if logout.StatusCode != 204 {
		t.Fatal(logout.StatusCode)
	}
	if authRequest(t, app, "GET", "/api/v1/session", "", "", "", cookie).StatusCode != 401 {
		t.Fatal("replayed logged-out cookie accepted")
	}
	login = authRequest(t, app, "POST", "/api/v1/session", `{"username":"alex","password":"correct-password"}`, "https://dashboard.test", "application/json", nil)
	cookie = login.Cookies()[0]
	a.Sessions.mu.Lock()
	entry := a.Sessions.sessions[cookie.Value]
	entry.expires = time.Now().Add(-time.Second)
	a.Sessions.sessions[cookie.Value] = entry
	a.Sessions.mu.Unlock()
	if authRequest(t, app, "GET", "/api/v1/session", "", "", "", cookie).StatusCode != 401 {
		t.Fatal("expired cookie accepted")
	}
}
func TestViewerAndCSRFProtection(t *testing.T) {
	_, app := accountTestApp(t, "viewer")
	login := authRequest(t, app, "POST", "/api/v1/session", `{"username":"alex","password":"correct-password"}`, "https://dashboard.test", "application/json", nil)
	if login.StatusCode != 204 {
		t.Fatal(login.StatusCode)
	}
	cookie := login.Cookies()[0]
	if authRequest(t, app, "POST", "/admin", "{}", "https://dashboard.test", "application/json", cookie).StatusCode != 403 {
		t.Fatal("viewer gained admin access")
	}
	for _, tc := range []struct{ origin, content string }{{"https://evil.test", "application/json"}, {"", "application/json"}, {"https://dashboard.test", "text/plain"}} {
		if authRequest(t, app, "DELETE", "/api/v1/session", "{}", tc.origin, tc.content, cookie).StatusCode != 403 {
			t.Fatal("unsafe logout accepted")
		}
		if authRequest(t, app, "POST", "/api/v1/session", `{"username":"alex","password":"correct-password"}`, tc.origin, tc.content, nil).StatusCode != 403 {
			t.Fatal("unsafe login accepted")
		}
	}
}
func TestGenericLoginFailuresAndThrottle(t *testing.T) {
	_, app := accountTestApp(t, "admin")
	var first string
	for i := 0; i < 6; i++ {
		username := "alex"
		if i%2 == 1 {
			username = "missing"
		}
		res := authRequest(t, app, "POST", "/api/v1/session", `{"username":"`+username+`","password":"wrong-password"}`, "https://dashboard.test", "application/json", nil)
		if i == 5 {
			if res.StatusCode != 429 {
				t.Fatal("unthrottled login")
			}
			continue
		}
		data, _ := io.ReadAll(res.Body)
		if i == 0 {
			first = string(data)
		}
		if res.StatusCode != 401 || string(data) != first {
			t.Fatal("credential failure reveals account existence")
		}
	}
}
func TestInvalidAccountConfigurationFailsClosed(t *testing.T) {
	a, app := accountTestApp(t, "admin")
	if err := os.Chmod(a.Config.AccountsFile, 0644); err != nil {
		t.Fatal(err)
	}
	a.RegisterSessionRoutes(app)
	if a.Sessions != nil {
		t.Fatal("public accounts file accepted")
	}
	a.Config.SessionSecret = "short"
	a.RegisterSessionRoutes(app)
	if a.Sessions != nil {
		t.Fatal("short signing secret accepted")
	}
}

func TestAccountFileValidation(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	for _, records := range [][]account{
		{},
		{{Username: "alex", PasswordHash: hash, Role: "owner"}},
		{{Username: "alex", PasswordHash: "plaintext", Role: "admin"}},
		{{Username: "alex", PasswordHash: hash, Role: "admin"}, {Username: "alex", PasswordHash: hash, Role: "viewer"}},
	} {
		path := filepath.Join(t.TempDir(), "accounts.json")
		data, _ := json.Marshal(records)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadAccounts(path); err == nil {
			t.Fatal("invalid account configuration accepted")
		}
	}
	if verifyPassword(hash, "wrong-password") || !verifyPassword(hash, "correct-password") {
		t.Fatal("password verification incorrect")
	}
	if _, err := HashPassword("short"); err == nil {
		t.Fatal("weak provisioned password accepted")
	}
}

func TestTamperedAndRestartedSessionsRejected(t *testing.T) {
	a, app := accountTestApp(t, "admin")
	login := authRequest(t, app, "POST", "/api/v1/session", `{"username":"alex","password":"correct-password"}`, "https://dashboard.test", "application/json", nil)
	cookie := login.Cookies()[0]
	original := cookie.Value
	cookie.Value = "x" + original
	if authRequest(t, app, "GET", "/api/v1/session", "", "", "", cookie).StatusCode != 401 {
		t.Fatal("tampered session accepted")
	}
	cookie.Value = original
	a.RegisterSessionRoutes(fiber.New())
	if authRequest(t, app, "GET", "/api/v1/session", "", "", "", cookie).StatusCode != 401 {
		t.Fatal("session survived restart")
	}
}

func TestLoginThrottleFailsClosedAtMemoryBound(t *testing.T) {
	a, app := accountTestApp(t, "admin")
	a.Sessions.mu.Lock()
	for i := 0; i < 1024; i++ {
		a.Sessions.attempts[fmt.Sprint(i)] = loginWindow{count: 1, expires: time.Now().Add(time.Minute)}
	}
	a.Sessions.mu.Unlock()
	res := authRequest(t, app, "POST", "/api/v1/session", `{"username":"alex","password":"correct-password"}`, "https://dashboard.test", "application/json", nil)
	if res.StatusCode != 429 {
		t.Fatal("full throttle table allowed new key")
	}
	if len(a.Sessions.attempts) != 1024 {
		t.Fatal("throttle table grew")
	}
}

func TestSessionLifecycleAuditedWhenStorageConfigured(t *testing.T) {
	a, app := accountTestApp(t, "admin")
	a.Config.DataPath = t.TempDir()
	login := authRequest(t, app, "POST", "/api/v1/session", `{"username":"alex","password":"correct-password"}`, "https://dashboard.test", "application/json", nil)
	if login.StatusCode != 204 {
		t.Fatal(login.StatusCode)
	}
	logout := authRequest(t, app, "DELETE", "/api/v1/session", "{}", "https://dashboard.test", "application/json", login.Cookies()[0])
	if logout.StatusCode != 204 {
		t.Fatal(logout.StatusCode)
	}
	data, err := os.ReadFile(filepath.Join(a.Config.DataPath, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"action":"login"`) || !strings.Contains(string(data), `"action":"logout"`) {
		t.Fatalf("missing lifecycle audit: %s", data)
	}
}

func TestSessionSecretIsNotALoginCredential(t *testing.T) {
	a := &API{Config: config.Config{SessionSecret: strings.Repeat("s", 32)}}
	app := fiber.New()
	app.Post("/api/v1/session", a.login)
	req := httptest.NewRequest("POST", "http://localhost/api/v1/session", strings.NewReader(`{"password":"`+strings.Repeat("s", 32)+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 503 {
		t.Fatalf("unconfigured login = %d, want 503", res.StatusCode)
	}
}
