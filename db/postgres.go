package db

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"

	"goquickactu/sources"
	"goquickactu/topics"
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

		if err := seedSources(conn); err != nil {
			poolErr = err
			return
		}

		if err := seedTopics(conn); err != nil {
			poolErr = err
			return
		}

		pool = conn
	})

	return pool, poolErr
}

func initSchema(conn *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sources (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL UNIQUE,
			political_leaning TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS keywords (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS topics (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS articles (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			link TEXT NOT NULL UNIQUE,
			description TEXT,
			source_id INTEGER REFERENCES sources(id),
			topic_id INTEGER REFERENCES topics(id),
			published_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS article_keywords (
			article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
			keyword_id INTEGER NOT NULL REFERENCES keywords(id) ON DELETE CASCADE,
			PRIMARY KEY (article_id, keyword_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles (published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_source_id ON articles (source_id)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_topic_id ON articles (topic_id)`,
	}

	for _, stmt := range statements {
		if _, err := conn.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

// seedSources keeps the "sources" table in sync with the application's
// canonical feed list (goquickactu/sources), so the two never drift apart.
func seedSources(conn *sql.DB) error {
	for _, s := range sources.FeedSources {
		if _, err := conn.Exec(`
			INSERT INTO sources (name, url, political_leaning)
			VALUES ($1, $2, $3)
			ON CONFLICT (name) DO UPDATE SET
				url = EXCLUDED.url,
				political_leaning = EXCLUDED.political_leaning
		`, s.Name, s.URL, s.PoliticalLeaning); err != nil {
			return err
		}
	}
	return nil
}

// seedTopics keeps the "topics" table in sync with the application's
// canonical taxonomy (goquickactu/topics), so the two never drift apart.
func seedTopics(conn *sql.DB) error {
	for _, name := range topics.Topics {
		if _, err := conn.Exec(`
			INSERT INTO topics (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
		`, name); err != nil {
			return err
		}
	}
	return nil
}
