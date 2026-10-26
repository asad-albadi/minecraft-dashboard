package httpapi

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"mime"
	"net/url"
	"strings"
	"sync"
	"time"
)

const sessionCookie = "minecraft_dashboard_session"
const sessionLifetime = 8 * time.Hour
const identityKey = "dashboard.identity"

type Identity struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}
type sessionEntry struct {
	identity Identity
	expires  time.Time
}
type loginWindow struct {
	count   int
	expires time.Time
}
type sessionManager struct {
	mu        sync.Mutex
	accounts  map[string]account
	sessions  map[string]sessionEntry
	attempts  map[string]loginWindow
	dummyHash string
}

func (a *API) RegisterSessionRoutes(app *fiber.App) {
	a.Sessions = nil
	if len(a.Config.SessionSecret) >= 32 {
		if accounts, err := loadAccounts(a.Config.AccountsFile); err == nil {
			dummy, err := HashPassword("dummy-account-password")
			if err == nil {
				a.Sessions = &sessionManager{accounts: accounts, sessions: make(map[string]sessionEntry), attempts: make(map[string]loginWindow), dummyHash: dummy}
			}
		}
	}
	app.Post("/api/v1/session", a.login)
	app.Get("/api/v1/session", a.requireSession, func(c *fiber.Ctx) error { return c.JSON(c.Locals(identityKey)) })
	app.Delete("/api/v1/session", a.requireSession, a.logout)
}
func (a *API) csrf(c *fiber.Ctx) bool {
	origin := a.Config.PublicOrigin
	if origin == "" {
		origin = c.Protocol() + "://" + c.Get("Host")
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return false
	}
	media, _, err := mime.ParseMediaType(c.Get("Content-Type"))
	return err == nil && media == "application/json" && c.Get("Origin") == origin
}
func (a *API) login(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	s := a.Sessions
	if s == nil {
		return c.Status(503).JSON(fiber.Map{"error": "dashboard access is not configured"})
	}
	if !a.csrf(c) {
		return c.Status(403).JSON(fiber.Map{"error": "request origin or content type rejected"})
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if len(c.Body()) > 4096 || json.Unmarshal(c.Body(), &req) != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, w := range s.attempts {
		if !now.Before(w.expires) {
			delete(s.attempts, key)
		}
	}
	ip := c.IP()
	w, exists := s.attempts[ip]
	if (!exists && len(s.attempts) >= 1024) || w.count >= 5 {
		c.Set("Retry-After", "60")
		return c.Status(429).JSON(fiber.Map{"error": "too many login attempts"})
	}
	if !exists {
		w.expires = now.Add(time.Minute)
	}
	w.count++
	s.attempts[ip] = w
	account, found := s.accounts[req.Username]
	hash := s.dummyHash
	if found {
		hash = account.PasswordHash
	}
	valid := verifyPassword(hash, req.Password)
	if !valid || !found {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}
	for key, entry := range s.sessions {
		if !now.Before(entry.expires) {
			delete(s.sessions, key)
		}
	}
	if len(s.sessions) >= 1024 {
		return c.Status(503).JSON(fiber.Map{"error": "session capacity reached"})
	}
	id := make([]byte, 32)
	if _, err := rand.Read(id); err != nil {
		return c.SendStatus(503)
	}
	token := base64.RawURLEncoding.EncodeToString(id)
	mac := hmac.New(sha256.New, []byte(a.Config.SessionSecret))
	mac.Write([]byte(token))
	token += "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	expires := now.Add(sessionLifetime)
	s.sessions[token] = sessionEntry{Identity{account.Username, account.Role}, expires}
	a.auditSession(account.Username, "login")
	c.Cookie(&fiber.Cookie{Name: sessionCookie, Value: token, Path: "/", HTTPOnly: true, Secure: a.Config.CookieSecure, SameSite: "Strict", Expires: expires, MaxAge: int(sessionLifetime.Seconds())})
	return c.SendStatus(204)
}
func (a *API) requireSession(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	if a.Sessions == nil {
		return c.Status(503).JSON(fiber.Map{"error": "dashboard access is not configured"})
	}
	token := c.Cookies(sessionCookie)
	parts := strings.Split(token, ".")
	valid := false
	if len(parts) == 2 {
		sig, err := base64.RawURLEncoding.DecodeString(parts[1])
		mac := hmac.New(sha256.New, []byte(a.Config.SessionSecret))
		mac.Write([]byte(parts[0]))
		valid = err == nil && hmac.Equal(sig, mac.Sum(nil))
	}
	a.Sessions.mu.Lock()
	entry, found := a.Sessions.sessions[token]
	if found && !time.Now().Before(entry.expires) {
		delete(a.Sessions.sessions, token)
		found = false
	}
	a.Sessions.mu.Unlock()
	if !valid || !found {
		return c.Status(401).JSON(fiber.Map{"error": "authentication required"})
	}
	if c.Method() != "GET" && c.Method() != "HEAD" && c.Method() != "OPTIONS" && !a.csrf(c) {
		return c.Status(403).JSON(fiber.Map{"error": "request origin or content type rejected"})
	}
	c.Locals(identityKey, entry.identity)
	c.Locals("username", entry.identity.Username)
	c.Locals("role", entry.identity.Role)
	c.Locals("sessionExpires", entry.expires)
	return c.Next()
}

// sessionActive allows long-lived event streams to stop after logout or expiry.
func (a *API) sessionActive(token string) bool {
	if a.Sessions == nil {
		return false
	}
	a.Sessions.mu.Lock()
	defer a.Sessions.mu.Unlock()
	entry, ok := a.Sessions.sessions[token]
	return ok && time.Now().Before(entry.expires)
}
func (a *API) requireAdmin(c *fiber.Ctx) error {
	identity, ok := c.Locals(identityKey).(Identity)
	if !ok || identity.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "administrator access required"})
	}
	return c.Next()
}
func (a *API) logout(c *fiber.Ctx) error {
	a.Sessions.mu.Lock()
	delete(a.Sessions.sessions, c.Cookies(sessionCookie))
	a.Sessions.mu.Unlock()
	if actor, ok := c.Locals("username").(string); ok {
		a.auditSession(actor, "logout")
	}
	c.Cookie(&fiber.Cookie{Name: sessionCookie, Value: "", Path: "/", HTTPOnly: true, Secure: a.Config.CookieSecure, SameSite: "Strict", Expires: time.Unix(1, 0), MaxAge: -1})
	return c.SendStatus(204)
}
