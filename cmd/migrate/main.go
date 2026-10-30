package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"otome-server/internal/shared/config"
	"otome-server/migrations"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dsn, err := mysql.ParseDSN(cfg.MySQLDSN)
	if err != nil {
		return fmt.Errorf("parse mysql dsn: %w", err)
	}
	// Migration files contain multiple statements.
	dsn.MultiStatements = true

	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	defer db.Close()

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}
	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		return fmt.Errorf("create migrate driver: %w", err)
	}
	migrator, err := migrate.NewWithInstance("iofs", source, "mysql", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		err = migrator.Up()
	case "down":
		err = migrator.Steps(-1)
	case "version":
		version, dirty, versionErr := migrator.Version()
		if versionErr != nil && !errors.Is(versionErr, migrate.ErrNilVersion) {
			return versionErr
		}
		fmt.Printf("version=%d dirty=%v\n", version, dirty)
		return nil
	default:
		return fmt.Errorf("unknown command %q (use up|down|version)", command)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("no change")
	} else {
		fmt.Printf("migrate %s done\n", command)
	}
	return nil
}
