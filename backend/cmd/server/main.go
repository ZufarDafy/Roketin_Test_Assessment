package main

import (
	"log"

	"minishop/internal/config"
	"minishop/internal/database"
	"minishop/internal/router"
)

func main() {
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

	r := router.New(db)
	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
