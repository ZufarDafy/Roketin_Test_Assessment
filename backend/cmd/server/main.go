package main

import (
	"log"
	"os"

	"minishop/internal/config"
	"minishop/internal/database"
	"minishop/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// Default release mode agar log bersih; override via GIN_MODE=debug.
	if os.Getenv(gin.EnvGinMode) == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg := config.Load()

	db, err := database.Connect(cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	if err := database.Seed(db); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	r := router.New(db, cfg.AllowedOrigins)
	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
