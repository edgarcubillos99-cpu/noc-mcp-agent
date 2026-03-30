package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv         string
	LogLevel       string
	MaxPingWorkers int
	MaxNmapWorkers int
	HTTPPort       string
	APIKey         string
}

var App Config

func Load() {
	App = Config{
		AppEnv:         getEnv("APP_ENV", "production"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MaxPingWorkers: getEnvAsInt("MAX_PING_WORKERS", 50),
		MaxNmapWorkers: getEnvAsInt("MAX_NMAP_WORKERS", 5),
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		APIKey:         getEnv("API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
