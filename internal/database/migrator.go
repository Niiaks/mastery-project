package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"mastery-project/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(cfg *config.Config) error {
	var migrationsPath string

	if cfg.ENV == "production" {
		migrationsPath = "file:///var/www/myapp/database/migrations"
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %v", err)
		}

		migrationsFolder := filepath.Join(cwd, "internal", "database", "migrations")
		migrationsFolder = filepath.ToSlash(migrationsFolder)

		migrationsPath = fmt.Sprintf("file://%s", migrationsFolder)
	}

	dbDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.DBUser,
		cfg.Database.DBPass,
		cfg.Database.DBHost,
		cfg.Database.DBPort,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	m, err := migrate.New(migrationsPath, dbDSN)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	log.Println("Migrations applied successfully!")
	return nil
}
