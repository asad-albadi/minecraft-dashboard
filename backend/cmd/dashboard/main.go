package main

import (
	"log"
	"minecraft-dashboard/backend/internal/config"
	"minecraft-dashboard/backend/internal/httpapi"
	"github.com/gofiber/fiber/v2"
)
func main() { cfg := config.Load(); app := fiber.New(fiber.Config{AppName:"Minecraft Dashboard"}); httpapi.New(cfg).Register(app); log.Fatal(app.Listen(cfg.ListenAddress)) }
