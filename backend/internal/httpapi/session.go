package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const sessionCookie = "minecraft_dashboard_session"
const sessionLifetime = 8 * time.Hour

func (a *API) login(c *fiber.Ctx) error {
	if a.Config.SessionSecret == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "dashboard access is not configured"})
	}
	var request struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if subtle.ConstantTimeCompare([]byte(request.Password), []byte(a.Config.SessionSecret)) != 1 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	expires := time.Now().Add(sessionLifetime)
	c.Cookie(&fiber.Cookie{Name: sessionCookie, Value: a.sessionToken(expires), Path: "/", HTTPOnly: true, SameSite: "Strict", Expires: expires, MaxAge: int(sessionLifetime.Seconds())})
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) requireSession(c *fiber.Ctx) error {
	if a.Config.SessionSecret == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "dashboard access is not configured"})
	}
	if !a.validSession(c.Cookies(sessionCookie)) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
	}
	return c.Next()
}

func (a *API) sessionToken(expires time.Time) string {
	payload := strconv.FormatInt(expires.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(a.Config.SessionSecret))
	_, _ = mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (a *API) validSession(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	expires, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().Unix() >= expires {
		return false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(a.Config.SessionSecret))
	_, _ = mac.Write([]byte(parts[0]))
	return hmac.Equal(signature, mac.Sum(nil))
}
