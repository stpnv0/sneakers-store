package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("POSTGRES_DSN")
	dir := flag.String("dir", "/app/migrations", "migrations directory")
	command := flag.String("command", "up", "goose command (up, down, status, reset, version)")
	flag.Parse()

	if dsn == "" {
		return fmt.Errorf("POSTGRES_DSN environment variable is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err := goose.Run(*command, db, *dir); err != nil {
		return fmt.Errorf("goose %s: %w", *command, err)
	}

	fmt.Println("migrations applied successfully")
	return nil
}
