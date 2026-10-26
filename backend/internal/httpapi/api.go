package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"minecraft-dashboard/backend/internal/backups"
	"minecraft-dashboard/backend/internal/capabilities"
	"minecraft-dashboard/backend/internal/config"
	"minecraft-dashboard/backend/internal/history"
	"minecraft-dashboard/backend/internal/logs"
	"minecraft-dashboard/backend/internal/ping"
	"minecraft-dashboard/backend/internal/players"
	"minecraft-dashboard/backend/internal/rcon"
	"minecraft-dashboard/backend/internal/world"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type API struct {
	Sessions     *sessionManager
	Actions      actionManager
	History      *history.Store
	historyError bool
	monitorMu    sync.Mutex
	lastStatus   fiber.Map
	lastStatusAt time.Time
	lastSuccess  time.Time
	worldCache   *world.Cache
	stop         chan struct{}
	stopOnce     sync.Once
	logMu        sync.RWMutex
	lines        []string
	logError     bool
	logSequence  uint64
	Config       config.Config
	Capabilities capabilities.Snapshot
	RCON         *rcon.Client
	Events       *eventBroker
}

func New(c config.Config) *API {
	if c.RCONTimeout <= 0 {
		c.RCONTimeout = 3 * time.Second
	}
	api := &API{Config: c, Capabilities: capabilities.SnapshotFor(c), RCON: rcon.New(c.RCONAddress, c.RCONPassword, c.RCONTimeout), Events: newEventBroker(25), worldCache: world.NewCache(c.PropertiesPath, c.WorldPath, 5*time.Minute), stop: make(chan struct{}), lines: []string{}}
	var err error
	storage := c.DataPath
	if storage != "" {
		if _, e := api.auditDirectory(); e != nil {
			storage = ""
			api.historyError = true
		}
	}
	api.History, err = history.NewStore(storage)
	if err != nil {
		api.historyError = true
		api.History, _ = history.NewStore("")
	}
	go api.pollEvents()
	return api
}
func (a *API) Register(app *fiber.App) {
	app.Get("/api/v1/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) })
	a.RegisterSessionRoutes(app)
	api := app.Group("/api/v1", a.requireSession)
	api.Get("/capabilities", func(c *fiber.Ctx) error {
		caps := a.Capabilities
		if c.Locals("role") != "admin" {
			caps.Chat = false
			caps.Console = false
		}
		return c.JSON(caps)
	})
	api.Get("/status", a.status)
	api.Get("/metrics", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"tps": nil, "memory": nil, "uptime": nil}) })
	api.Get("/players", a.players)
	api.Get("/player-profiles", a.playerProfiles)
	api.Get("/world", a.world)
	api.Get("/logs", a.logs)
	api.Get("/settings", func(c *fiber.Ctx) error { return c.JSON(a.settingsPayload()) })
	api.Get("/backups", a.backups)
	api.Get("/audit", a.requireAdmin, a.auditLog)
	api.Get("/activity", func(c *fiber.Ctx) error { return c.JSON(a.History.Events()) })
	api.Get("/chat", func(c *fiber.Ctx) error { return c.JSON(a.History.Chat()) })
	api.Post("/chat/messages", a.requireAdmin, a.sendChat)
	api.Post("/console/commands", a.requireAdmin, a.runConsole)
	api.Post("/server/:operation", a.requireAdmin, a.unavailable("lifecycle"))
	api.Get("/events", a.events)
}
func (a *API) status(c *fiber.Ctx) error {
	return c.JSON(a.statusPayload(c.Context()))
}
func (a *API) statusPayload(ctx context.Context) fiber.Map {
	a.monitorMu.Lock()
	defer a.monitorMu.Unlock()
	if a.lastStatus != nil && time.Since(a.lastStatusAt) < 15*time.Second {
		return a.lastStatus
	}
	s := fiber.Map{"state": "unknown", "rcon": "Unavailable", "detail": "Game status is not configured"}
	port := a.Config.MinecraftPort
	if port == "" {
		port = "25565"
	}
	if a.Config.MinecraftHost != "" {
		data, err := ping.Query(ctx, net.JoinHostPort(a.Config.MinecraftHost, port))
		if err == nil {
			s["state"] = "online"
			s["version"] = data.Version
			s["players"] = data.Online
			s["maxPlayers"] = data.Max
			s["description"] = data.Description
			s["latencyMs"] = data.LatencyMS
			s["detail"] = "Java status ping reachable"
		} else {
			s["detail"] = "Game status ping unreachable; server state is unknown"
		}
	}
	if a.Config.RCONAddress != "" && a.Config.RCONPassword != "" {
		response, err := a.RCON.Command(ctx, "list")
		if err == nil {
			n, max, _, ok := rcon.ParseListResponse(response)
			if ok {
				s["state"] = "online"
				s["rcon"] = "Connected"
				s["players"] = n
				s["maxPlayers"] = max
				if s["version"] == nil {
					s["detail"] = "Read-only RCON reachable; game ping unavailable"
				}
			}
		} else {
			s["rcon"] = "Unreachable"
		}
	}
	now := time.Now().UTC()
	if s["state"] == "online" {
		a.lastSuccess = now
	}
	if !a.lastSuccess.IsZero() {
		s["lastSuccess"] = a.lastSuccess.Format(time.RFC3339)
	}
	s["updatedAt"] = now.Format(time.RFC3339)
	a.lastStatus = a.withVersion(s)
	a.lastStatusAt = now
	return a.lastStatus
}

func (a *API) Close() {
	a.stopOnce.Do(func() {
		if a.stop != nil {
			close(a.stop)
		}
	})
}
func (a *API) settingsPayload() fiber.Map {
	name := a.Config.ServerName
	if name == "" {
		name = "Minecraft Server"
	}
	port := a.Config.MinecraftPort
	if port == "" {
		port = "25565"
	}
	pathState := func(path string) string {
		if path == "" {
			return "Not configured"
		}
		if _, err := os.Stat(path); err != nil {
			return "Unavailable"
		}
		return "Configured (read-only)"
	}
	historyState := "Memory only (restart clears history)"
	if a.historyError {
		historyState = "Persistent storage unavailable; memory-only fallback"
	} else if a.History.Persistent() {
		historyState = "Persistent bounded history"
	}
	return fiber.Map{"serverName": name, "host": a.Config.MinecraftHost, "port": port, "integrations": fiber.Map{"History storage": historyState, "Log source": pathState(a.Config.LogPath), "World files": pathState(a.Config.WorldPath), "Backup catalog": pathState(a.Config.BackupPath), "RCON configuration": func() string {
		if a.Config.RCONAddress != "" && a.Config.RCONPassword != "" {
			return "Configured; live status on Overview"
		}
		return "Not configured"
	}(), "Container metrics": "Unavailable — restricted metrics adapter required", "Lifecycle": "Disabled — separate authorized control service required"}}
}
func (a *API) backups(c *fiber.Ctx) error {
	if a.Config.BackupPath == "" {
		return c.Status(503).JSON(fiber.Map{"error": "Backup directory is not configured. Mount an existing backup catalog read-only."})
	}
	entries, err := backups.List(a.Config.BackupPath)
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "Backup catalog unavailable or exceeds the safe listing limit"})
	}
	return c.JSON(entries)
}

func (a *API) withVersion(payload fiber.Map) fiber.Map {
	if version := logs.ServerVersion(a.logTail()); version != "" {
		payload["version"] = version
	}
	return payload
}

func (a *API) events(c *fiber.Ctx) error {
	token := string([]byte(c.Cookies(sessionCookie)))
	subscriber, cancel, ok := a.Events.subscribe()
	if !ok {
		c.Set(fiber.HeaderRetryAfter, "30")
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many event connections"})
	}
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		defer cancel()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		if !writeEvent(writer, "connection.state", fiber.Map{"state": "connected"}) {
			return
		}
		if !writeEvent(writer, "console.snapshot", a.consoleSnapshot()) {
			return
		}
		for {
			select {
			case <-a.stop:
				return
			case <-ticker.C:
				if !a.sessionActive(token) || !writeEvent(writer, "connection.state", fiber.Map{"state": "connected"}) {
					return
				}
			case event, ok := <-subscriber:
				if !ok || !a.sessionActive(token) || !writeEvent(writer, event.name, event.payload) {
					return
				}
			}
		}
	})
	return nil
}

func (a *API) logTail() []string {
	a.logMu.RLock()
	defer a.logMu.RUnlock()
	return append([]string{}, a.lines...)
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
		return c.Status(503).JSON(fiber.Map{"error": "online player list unavailable"})
	}
	_, _, names, ok := rcon.ParseListResponse(response)
	if !ok {
		return c.Status(503).JSON(fiber.Map{"error": "online player list unavailable"})
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
	profiles, err := players.ReadProfiles(a.Config.UserCachePath, a.Config.OpsPath, a.Config.WorldPath, a.Config.LogDirectory)
	if err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "player profiles unavailable"})
	}
	return c.JSON(profiles)
}
func (a *API) world(c *fiber.Ctx) error {
	if a.Config.PropertiesPath == "" || a.Config.WorldPath == "" {
		return c.JSON(fiber.Map{})
	}
	metadata, err := a.worldCache.Get()
	if err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "world metadata unavailable"})
	}
	if response, err := a.RCON.Command(c.Context(), "time query daytime"); err == nil {
		if day, clock, ok := world.ParseMinecraftTime(response); ok {
			metadata.Day, metadata.Time = day, clock
		}
	}
	if response, err := a.RCON.Command(c.Context(), "time query day"); err == nil {
		if day, ok := world.ParseMinecraftDay(response); ok {
			metadata.Day = day
		} else if _, clock, ok := world.ParseMinecraftTime(response); ok {
			// Newer servers return timeline ticks instead of a world-age count.
			metadata.Time = clock
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
	a.logMu.RLock()
	failed := a.logError
	a.logMu.RUnlock()
	if failed {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "log unavailable"})
	}
	return c.JSON(a.logTail())
}
func (a *API) unavailable(name string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": name + " capability is not configured"})
	}
}
