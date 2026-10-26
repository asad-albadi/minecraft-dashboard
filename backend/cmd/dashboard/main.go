package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"minecraft-dashboard/backend/internal/config"
	"minecraft-dashboard/backend/internal/httpapi"
)

func main() {
	cfg := config.Load()
	app := fiber.New(fiber.Config{AppName: "Minecraft Dashboard", BodyLimit: 8192})
	api := httpapi.New(cfg)
	defer api.Close()
	api.Register(app)
	log.Fatal(app.Listen(cfg.ListenAddress))
}
