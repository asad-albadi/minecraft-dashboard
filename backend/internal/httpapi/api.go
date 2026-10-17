package httpapi

import (
	"net/http"
	"minecraft-dashboard/backend/internal/capabilities"
	"minecraft-dashboard/backend/internal/chat"
	"minecraft-dashboard/backend/internal/config"
	"github.com/gofiber/fiber/v2"
)

type API struct { Config config.Config; Capabilities capabilities.Snapshot }
func New(c config.Config) *API { return &API{Config:c, Capabilities:capabilities.SnapshotFor(c)} }
func (a *API) Register(app *fiber.App) {
	app.Get("/api/v1/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok":true}) })
	app.Get("/api/v1/capabilities", func(c *fiber.Ctx) error { return c.JSON(a.Capabilities) })
	app.Get("/api/v1/status", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"state":"unknown","players":0,"maxPlayers":0}) })
	app.Get("/api/v1/metrics", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"tps":nil,"memory":nil,"uptime":nil}) })
	app.Get("/api/v1/players", func(c *fiber.Ctx) error { return c.JSON([]fiber.Map{}) })
	app.Get("/api/v1/world", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{}) })
	app.Get("/api/v1/logs", func(c *fiber.Ctx) error { return c.JSON([]fiber.Map{}) })
	app.Post("/api/v1/chat/messages", a.chat)
	app.Post("/api/v1/console/commands", a.unavailable("console"))
	app.Post("/api/v1/server/:operation", a.unavailable("lifecycle"))
	app.Get("/api/v1/events", func(c *fiber.Ctx) error { return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error":"event stream unavailable until a log adapter is configured"}) })
}
func (a *API) chat(c *fiber.Ctx) error { if !a.Capabilities.Chat { return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error":"chat capability is not configured"}) }; var request struct{ Message string `json:"message"` }; if err := c.BodyParser(&request); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }; command, err := chat.TellrawCommand(request.Message); if err != nil { return c.Status(422).JSON(fiber.Map{"error":err.Error()}) }; _ = command // RCON adapter is intentionally wired only after its private transport is configured.
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error":"RCON delivery adapter is not yet available"}) }
func (a *API) unavailable(name string) fiber.Handler { return func(c *fiber.Ctx) error { return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error":name+" capability is not configured"}) } }
