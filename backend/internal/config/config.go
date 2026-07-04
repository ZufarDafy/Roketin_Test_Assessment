package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN          string
	Port           string
	AllowedOrigins []string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		DBDSN:          mustGetEnv("DB_DSN"),
		Port:           mustGetEnv("PORT"),
		AllowedOrigins: splitCSV(mustGetEnv("ALLOWED_ORIGINS")),
	}
}

func mustGetEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	log.Fatalf("config: required env var %s is not set", key)
	return ""
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
