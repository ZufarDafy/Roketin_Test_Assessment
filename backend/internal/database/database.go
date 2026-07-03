package database

import (
	"minishop/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Migrate membuat tabel via AutoMigrate lalu menjalankan raw SQL untuk
// hal yang tidak bisa dibuat GORM: extension pg_trgm dan GIN trigram index
// (mempercepat search ILIKE '%keyword%' yang tidak bisa memakai B-tree
// karena leading wildcard). Semua statement idempotent.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		return err
	}

	stmts := []string{
		`CREATE EXTENSION IF NOT EXISTS pg_trgm`,
		`CREATE INDEX IF NOT EXISTS idx_products_name_trgm
			ON products USING GIN (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_products_category_id
			ON products (category_id)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}
