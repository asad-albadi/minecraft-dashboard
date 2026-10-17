package config

import ("os"; "time")

type Config struct { ListenAddress, MinecraftHost, RCONAddress, RCONPassword, LogPath, WorldPath, BackupPath, ContainerName, SessionSecret string; RCONTimeout time.Duration }
func Load() Config { return Config{ListenAddress: value("DASHBOARD_LISTEN", ":8080"), MinecraftHost: os.Getenv("MINECRAFT_HOST"), RCONAddress: os.Getenv("RCON_ADDRESS"), RCONPassword: os.Getenv("RCON_PASSWORD"), LogPath: os.Getenv("MINECRAFT_LOG_PATH"), WorldPath: os.Getenv("MINECRAFT_WORLD_PATH"), BackupPath: os.Getenv("DASHBOARD_BACKUP_PATH"), ContainerName: os.Getenv("MINECRAFT_CONTAINER_NAME"), SessionSecret: os.Getenv("SESSION_SECRET"), RCONTimeout: 5*time.Second} }
func value(key, fallback string) string { if v := os.Getenv(key); v != "" { return v }; return fallback }
