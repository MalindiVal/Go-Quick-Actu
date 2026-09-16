package db

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	pool     *sql.DB
	poolOnce sync.Once
	poolErr  error
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func dsn() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	name := getEnv("DB_NAME", "quickactu")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, name, sslmode)
}

// Connect returns a shared, lazily-initialized PostgreSQL connection pool,
// creating the schema on first use.
func Connect() (*sql.DB, error) {
	poolOnce.Do(func() {
		conn, err := sql.Open("postgres", dsn())
		if err != nil {
			poolErr = err
			return
		}

		if err := conn.Ping(); err != nil {
			poolErr = err
			return
		}

		if err := initSchema(conn); err != nil {
			poolErr = err
			return
		}

		pool = conn
	})

	return pool, poolErr
}

func initSchema(conn *sql.DB) error {
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			link TEXT NOT NULL UNIQUE,
			description TEXT,
			source TEXT,
			published_at TIMESTAMPTZ,
			keywords TEXT[]
		)
	`)
	return err
}
