package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	var (
		dbURL          = flag.String("db-url", getEnv("DATABASE_URL", "postgresql://sanjar:npg_oMOWeyCXh7a0@ep-young-glitter-a2q6k58c-pooler.eu-central-1.aws.neon.tech/my-db?sslmode=require"), "Database connection URL")
		migrationsPath = flag.String("migrations-path", getEnv("MIGRATIONS_PATH", "migrations"), "Path to migration files")
		command        = flag.String("command", "", "Migration command (up, down, version)")
		steps          = flag.Int("steps", 0, "Number of migration steps (0 means all)")
	)
	flag.Parse()

	if *command == "" {
		flag.Usage()
		os.Exit(1)
	}

	sourceURL := fmt.Sprintf("file://%s", *migrationsPath)
	m, err := migrate.New(sourceURL, *dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch *command {
	case "up":
		if *steps > 0 {
			if err := m.Steps(*steps); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Failed to apply %d migrations: %v", *steps, err)
			}
			log.Printf("Applied %d migrations", *steps)
		} else {
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Failed to apply all migrations: %v", err)
			}
			log.Println("Applied all migrations")
		}
	case "down":
		if *steps > 0 {
			if err := m.Steps(-(*steps)); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Failed to rollback %d migrations: %v", *steps, err)
			}
			log.Printf("Rolled back %d migrations", *steps)
		} else {
			if err := m.Down(); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Failed to rollback all migrations: %v", err)
			}
			log.Println("Rolled back all migrations")
		}
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		log.Printf("Current migration version: %d (dirty: %v)", version, dirty)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
