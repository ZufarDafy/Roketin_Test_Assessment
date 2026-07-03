package config

import "os"

type Config struct {
	DBDSN string
	Port  string
}

func Load() Config {
	return Config{
		DBDSN: getEnv("DB_DSN", "host=localhost user=minishop password=minishop dbname=minishop port=5432 sslmode=disable"),
		Port:  getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
