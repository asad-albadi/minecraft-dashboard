package httpapi

import (
	"net/http"
	"minecraft-dashboard/backend/internal/capabilities"
	"minecraft-dashboard/backend/internal/chat"
	"minecraft-dashboard/backend/internal/config"
	"minecraft-dashboard/backend/internal/rcon"
	"minecraft-dashboard/backend/internal/logs"
	"minecraft-dashboard/backend/internal/world"
	"github.com/gofiber/fiber/v2"
)

type API struct { Config config.Config; Capabilities capabilities.Snapshot; RCON *rcon.Client }
func New(c config.Config) *API { return &API{Config:c, Capabilities:capabilities.SnapshotFor(c), RCON:rcon.New(c.RCONAddress,c.RCONPassword,c.RCONTimeout)} }
func (a *API) Register(app *fiber.App) {
	app.Get("/api/v1/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok":true}) })
	app.Get("/api/v1/capabilities", func(c *fiber.Ctx) error { return c.JSON(a.Capabilities) })
	app.Get("/api/v1/status", a.status)
	app.Get("/api/v1/metrics", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"tps":nil,"memory":nil,"uptime":nil}) })
	app.Get("/api/v1/players", a.players)
	app.Get("/api/v1/world", a.world)
	app.Get("/api/v1/logs", a.logs)
	app.Post("/api/v1/chat/messages", a.chat)
	app.Post("/api/v1/console/commands", a.unavailable("console"))
	app.Post("/api/v1/server/:operation", a.unavailable("lifecycle"))
	app.Get("/api/v1/events", func(c *fiber.Ctx) error { return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error":"event stream unavailable until a log adapter is configured"}) })
}
func (a *API) status(c *fiber.Ctx) error { response, err := a.RCON.Command(c.Context(), "list"); if err != nil { return c.JSON(fiber.Map{"state":"offline","players":0,"maxPlayers":0,"detail":"RCON unreachable"}) }; players, max, _, ok := rcon.ParseListResponse(response); if !ok { return c.JSON(fiber.Map{"state":"unknown","players":0,"maxPlayers":0,"detail":"unrecognized RCON list response"}) }; return c.JSON(fiber.Map{"state":"online","players":players,"maxPlayers":max}) }
func (a *API) players(c *fiber.Ctx) error { response, err := a.RCON.Command(c.Context(), "list"); if err != nil { return c.JSON([]fiber.Map{}) }; _, _, names, ok := rcon.ParseListResponse(response); if !ok { return c.JSON([]fiber.Map{}) }; result := make([]fiber.Map, 0, len(names)); for _, name := range names { result = append(result, fiber.Map{"id":name,"name":name}) }; return c.JSON(result) }
func (a *API) world(c *fiber.Ctx) error { if a.Config.PropertiesPath == "" || a.Config.WorldPath == "" { return c.JSON(fiber.Map{}) }; metadata, err := world.ReadMetadata(a.Config.PropertiesPath, a.Config.WorldPath); if err != nil { return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error":"world metadata unavailable"}) }; return c.JSON(metadata) }
func (a *API) logs(c *fiber.Ctx) error { if a.Config.LogPath == "" { return c.JSON([]string{}) }; lines, err := logs.Tail(a.Config.LogPath, 100); if err != nil { return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error":"log unavailable"}) }; return c.JSON(lines) }
func (a *API) chat(c *fiber.Ctx) error { if !a.Capabilities.Chat { return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error":"chat capability is not configured"}) }; var request struct{ Message string `json:"message"` }; if err := c.BodyParser(&request); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request"}) }; command, err := chat.TellrawCommand(request.Message); if err != nil { return c.Status(422).JSON(fiber.Map{"error":err.Error()}) }; _ = command // RCON adapter is intentionally wired only after its private transport is configured.
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error":"RCON delivery adapter is not yet available"}) }
func (a *API) unavailable(name string) fiber.Handler { return func(c *fiber.Ctx) error { return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error":name+" capability is not configured"}) } }
