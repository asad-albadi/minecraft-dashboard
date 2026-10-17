package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"minecraft-dashboard/backend/internal/capabilities"
	"minecraft-dashboard/backend/internal/chat"
	"minecraft-dashboard/backend/internal/config"
	"minecraft-dashboard/backend/internal/logs"
	"minecraft-dashboard/backend/internal/players"
	"minecraft-dashboard/backend/internal/rcon"
	"minecraft-dashboard/backend/internal/world"
	"net/http"
	"time"
)

type API struct {
	Config       config.Config
	Capabilities capabilities.Snapshot
	RCON         *rcon.Client
}

func New(c config.Config) *API {
	return &API{Config: c, Capabilities: capabilities.SnapshotFor(c), RCON: rcon.New(c.RCONAddress, c.RCONPassword, c.RCONTimeout)}
}
func (a *API) Register(app *fiber.App) {
	app.Get("/api/v1/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) })
	app.Get("/api/v1/capabilities", func(c *fiber.Ctx) error { return c.JSON(a.Capabilities) })
	app.Get("/api/v1/status", a.status)
	app.Get("/api/v1/metrics", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"tps": nil, "memory": nil, "uptime": nil}) })
	app.Get("/api/v1/players", a.players)
	app.Get("/api/v1/player-profiles", a.playerProfiles)
	app.Get("/api/v1/world", a.world)
	app.Get("/api/v1/logs", a.logs)
	app.Post("/api/v1/chat/messages", a.chat)
	app.Post("/api/v1/console/commands", a.unavailable("console"))
	app.Post("/api/v1/server/:operation", a.unavailable("lifecycle"))
	app.Get("/api/v1/events", a.events)
}
func (a *API) status(c *fiber.Ctx) error {
	return c.JSON(a.statusPayload(c.Context()))
}
func (a *API) statusPayload(ctx context.Context) fiber.Map {
	response, err := a.RCON.Command(ctx, "list")
	if err != nil {
		return a.withVersion(fiber.Map{"state": "offline", "players": 0, "maxPlayers": 0, "detail": "RCON unreachable"})
	}
	players, max, _, ok := rcon.ParseListResponse(response)
	if !ok {
		return a.withVersion(fiber.Map{"state": "unknown", "players": 0, "maxPlayers": 0, "detail": "unrecognized RCON list response"})
	}
	return a.withVersion(fiber.Map{"state": "online", "players": players, "maxPlayers": max})
}

func (a *API) withVersion(payload fiber.Map) fiber.Map {
	if version := logs.ServerVersion(a.logTail()); version != "" {
		payload["version"] = version
	}
	return payload
}

func (a *API) events(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		previous := a.logTail()
		if !writeEvent(writer, "connection.state", fiber.Map{"state": "connected"}) || !writeEvent(writer, "server.status", a.statusPayload(context.Background())) {
			return
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for tick := 1; ; tick++ {
			<-ticker.C
			current := a.logTail()
			for _, line := range logs.NewLines(previous, current) {
				if !writeEvent(writer, "console.line", fiber.Map{"line": line}) {
					return
				}
			}
			previous = current
			if tick%15 == 0 && !writeEvent(writer, "server.status", a.statusPayload(context.Background())) {
				return
			}
		}
	})
	return nil
}

func (a *API) logTail() []string {
	if a.Config.LogPath == "" {
		return nil
	}
	lines, err := logs.Tail(a.Config.LogPath, 100)
	if err != nil {
		return nil
	}
	return lines
}

func writeEvent(writer *bufio.Writer, name string, payload any) bool {
	data, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	if _, err = fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", name, data); err != nil {
		return false
	}
	return writer.Flush() == nil
}
func (a *API) players(c *fiber.Ctx) error {
	response, err := a.RCON.Command(c.Context(), "list")
	if err != nil {
		return c.JSON([]fiber.Map{})
	}
	_, _, names, ok := rcon.ParseListResponse(response)
	if !ok {
		return c.JSON([]fiber.Map{})
	}
	result := make([]fiber.Map, 0, len(names))
	for _, name := range names {
		result = append(result, fiber.Map{"id": name, "name": name})
	}
	return c.JSON(result)
}
func (a *API) playerProfiles(c *fiber.Ctx) error {
	if a.Config.UserCachePath == "" || a.Config.WorldPath == "" {
		return c.JSON([]players.Profile{})
	}
	profiles, err := players.ReadProfiles(a.Config.UserCachePath, a.Config.OpsPath, a.Config.WorldPath)
	if err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "player profiles unavailable"})
	}
	return c.JSON(profiles)
}
func (a *API) world(c *fiber.Ctx) error {
	if a.Config.PropertiesPath == "" || a.Config.WorldPath == "" {
		return c.JSON(fiber.Map{})
	}
	metadata, err := world.ReadMetadata(a.Config.PropertiesPath, a.Config.WorldPath)
	if err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "world metadata unavailable"})
	}
	if response, err := a.RCON.Command(c.Context(), "time query day"); err == nil {
		if day, clock, ok := world.ParseMinecraftTime(response); ok {
			metadata.Day, metadata.Time = day, clock
		}
	}
	if response, err := a.RCON.Command(c.Context(), "weather query"); err == nil {
		if weather, ok := world.ParseWeather(response); ok {
			metadata.Weather = weather
		}
	}
	return c.JSON(metadata)
}
func (a *API) logs(c *fiber.Ctx) error {
	if a.Config.LogPath == "" {
		return c.JSON([]string{})
	}
	lines, err := logs.Tail(a.Config.LogPath, 100)
	if err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "log unavailable"})
	}
	return c.JSON(lines)
}
func (a *API) chat(c *fiber.Ctx) error {
	if !a.Capabilities.Chat {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "chat capability is not configured"})
	}
	var request struct {
		Message string `json:"message"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	command, err := chat.TellrawCommand(request.Message)
	if err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}
	_ = command // RCON adapter is intentionally wired only after its private transport is configured.
	return c.Status(http.StatusNotImplemented).JSON(fiber.Map{"error": "RCON delivery adapter is not yet available"})
}
func (a *API) unavailable(name string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": name + " capability is not configured"})
	}
}
