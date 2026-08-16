package infra

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() *sql.DB {
	sslMode := os.Getenv("SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	// This app uses Postgres.
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		sslMode,
	)

	dsnOverride := os.Getenv("DB_DSN_OVERRIDE")
	if dsnOverride != "" {
		dsn = dsnOverride
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Test ping
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
