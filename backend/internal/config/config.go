package config

import (
	"os"
	"time"
)

type Config struct {
	EnableChat, EnableConsole                                                                                                                                                   bool
	AccountsFile, PublicOrigin                                                                                                                                                  string
	ServerName, MinecraftPort, DataPath                                                                                                                                         string
	CookieSecure                                                                                                                                                                bool
	ListenAddress, MinecraftHost, RCONAddress, RCONPassword, LogPath, LogDirectory, WorldPath, PropertiesPath, UserCachePath, OpsPath, BackupPath, ContainerName, SessionSecret string
	RCONTimeout                                                                                                                                                                 time.Duration
}

func Load() Config {
	return loadWithAccounts()
}
func loadWithAccounts() Config {
	c := loadBase()
	c.AccountsFile = os.Getenv("DASHBOARD_ACCOUNTS_FILE")
	c.PublicOrigin = os.Getenv("DASHBOARD_PUBLIC_ORIGIN")
	c.CookieSecure = os.Getenv("DASHBOARD_COOKIE_SECURE") != "false"
	c.ServerName = value("DASHBOARD_SERVER_NAME", "Minecraft Server")
	c.MinecraftPort = value("MINECRAFT_PORT", "25565")
	c.DataPath = os.Getenv("DASHBOARD_DATA_PATH")
	c.EnableChat = os.Getenv("DASHBOARD_ENABLE_CHAT") == "true"
	c.EnableConsole = os.Getenv("DASHBOARD_ENABLE_CONSOLE") == "true"
	return c
}
func loadBase() Config {
	return Config{ListenAddress: value("DASHBOARD_LISTEN", ":8080"), MinecraftHost: os.Getenv("MINECRAFT_HOST"), RCONAddress: os.Getenv("RCON_ADDRESS"), RCONPassword: os.Getenv("RCON_PASSWORD"), LogPath: os.Getenv("MINECRAFT_LOG_PATH"), LogDirectory: os.Getenv("MINECRAFT_LOG_DIRECTORY"), WorldPath: os.Getenv("MINECRAFT_WORLD_PATH"), PropertiesPath: os.Getenv("MINECRAFT_PROPERTIES_PATH"), UserCachePath: os.Getenv("MINECRAFT_USERCACHE_PATH"), OpsPath: os.Getenv("MINECRAFT_OPS_PATH"), BackupPath: os.Getenv("DASHBOARD_BACKUP_PATH"), ContainerName: os.Getenv("MINECRAFT_CONTAINER_NAME"), SessionSecret: os.Getenv("SESSION_SECRET"), RCONTimeout: 5 * time.Second}
}
func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
