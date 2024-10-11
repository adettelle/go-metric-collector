package migrator

import (
	"database/sql"
	"embed"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const migrationDir = "migration"

//go:embed migration/*.sql
var MigrationsFS embed.FS

func MustApplyMigrations(dbParams string) {
	err := ApplyMigrations(dbParams)
	if err != nil {
		log.Fatalf("error applying migration: %v", err)
	}
}

func ApplyMigrations(dbParams string) error {
	srcDriver, err := iofs.New(MigrationsFS, migrationDir)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", dbParams)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithInstance("migration_embeded_sql_files", srcDriver, "psql_db", driver)
	if err != nil {
		return err
	}

	defer func() {
		migrator.Close()
	}()

	if err = migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Printf("Migrations applied")
	return nil
}

func ResetMigrations(dbParams string) error {
	srcDriver, err := iofs.New(MigrationsFS, migrationDir)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", dbParams)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithInstance("migration_embeded_sql_files", srcDriver, "psql_db", driver)
	if err != nil {
		return err
	}

	defer func() {
		migrator.Close()
	}()

	if err = migrator.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Printf("Migrations dropped")
	return nil
}
