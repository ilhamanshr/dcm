package database

import (
	"controller-service/internal/config"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func MigrateAll(db *sql.DB) error {
	slog.Info("Migrating pending migrations...")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("Error on initiating postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://./internal/database", config.Load().DBURL, driver)
	if err != nil {
		return fmt.Errorf("Error on NewWithDatabaseInstance(): %v", err)
	}

	err = m.Up()
	if err != nil && err == migrate.ErrNoChange {
		slog.Info("There is no pending migration")
		return nil
	}

	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("Error when running migrate up: %v", err)
	}

	slog.Info("All migration has been migrated successfully!")

	return nil
}
