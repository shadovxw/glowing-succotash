package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	return err
}

// Migrate runs all .sql files in the migrations directory in order.
func Migrate(migrationsDir string) error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	_, err = sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		var count int
		row := sqlDB.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE filename = $1`, name)
		if err := row.Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}

		if _, err := sqlDB.Exec(`INSERT INTO schema_migrations (filename) VALUES ($1)`, name); err != nil {
			return err
		}

		fmt.Printf("applied migration: %s\n", name)
	}
	return nil
}
