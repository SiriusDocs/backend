package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	var dbURL, migrationsPath, migrationsTable, level string

	flag.StringVar(&dbURL, "db-url", "", "PostgreSQL connection URL")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.StringVar(&level, "level", "", "up or down the migration")
	flag.Parse()

	if dbURL == "" {
		panic("storage-path is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}
	if level == "" {
		panic("it is impossible to understand whether to up or down migrations")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(fmt.Sprintf("failed to open PostgreSQL database: %v", err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	goose.SetTableName(migrationsTable)

	switch level {
	case "up":
		if err := goose.Up(db, migrationsPath); err != nil {
			panic(fmt.Sprintf("failed to up migrations: %v", err))
		}
	case "down":
		if err := goose.Down(db, migrationsPath); err != nil {
			panic(fmt.Sprintf("failed to down migrations: %v", err))
		}
	default:
		panic("it is impossible to understand whether to up or down migrations")
	}
}
