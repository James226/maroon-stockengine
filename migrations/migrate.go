package migrations

import (
	"database/sql"
	_ "path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(db *sql.DB, version uint) error {
	instance, err := WithInstance(db, &Config{
		DatabaseName:    "master",
		SchemaName:      "dbo",
		MigrationsTable: "migrations",
	})
	if err != nil {
		return err
	}

	databaseInstance, err := migrate.NewWithDatabaseInstance("file://migrations/sql", "master", instance)
	if err != nil {
		return err
	}

	err = databaseInstance.Migrate(version)
	if err != nil {
		return err
	}

	return nil
}
