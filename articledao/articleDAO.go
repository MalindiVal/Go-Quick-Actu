package articledao

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"goquickactu/db"
	"goquickactu/models"
)

// SaveArticle inserts or updates an article.
//
// analyzed tells whether article.Keywords / article.Topic come from a
// successful AI analysis. When it is false (the analysis failed), only the
// RSS fields are written: an existing article keeps its topic and keywords
// instead of having them wiped by a transient AI failure.
func SaveArticle(article models.Article, analyzed bool) error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}

	ctx := context.Background()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var sourceID int
	if err := tx.QueryRowContext(ctx, `SELECT id FROM sources WHERE name = $1`, article.Source).Scan(&sourceID); err != nil {
		return err
	}

	var topicID sql.NullInt64
	if article.Topic != "" {
		var id int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM topics WHERE name = $1`, article.Topic).Scan(&id)
		switch {
		case err == nil:
			topicID = sql.NullInt64{Int64: id, Valid: true}
		case err == sql.ErrNoRows:
			// Unknown topic (e.g. a stale AI classification): leave the
			// article uncategorized rather than failing the whole save.
		default:
			return err
		}
	}

	var articleID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO articles (title, link, description, source_id, topic_id, published_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (link) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			source_id = EXCLUDED.source_id,
			topic_id = CASE WHEN $7::boolean THEN EXCLUDED.topic_id ELSE articles.topic_id END,
			published_at = EXCLUDED.published_at
		RETURNING id
	`, article.Title, article.Link, article.Description, sourceID, topicID, article.PublishedAt, analyzed).Scan(&articleID)
	if err != nil {
		return err
	}

	if !analyzed {
		return tx.Commit()
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM article_keywords WHERE article_id = $1`, articleID); err != nil {
		return err
	}

	for _, keyword := range article.Keywords {
		var keywordID int
		err = tx.QueryRowContext(ctx, `
			INSERT INTO keywords (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, keyword).Scan(&keywordID)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO article_keywords (article_id, keyword_id)
			VALUES ($1, $2)
			ON CONFLICT (article_id, keyword_id) DO NOTHING
		`, articleID, keywordID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAnalyzedLinks returns the links of the articles that already have
// keywords, i.e. that a previous aggregation analyzed successfully, so the
// aggregation can skip them instead of paying for a new AI call.
func GetAnalyzedLinks() (map[string]bool, error) {
	conn, err := db.Connect()
	if err != nil {
		return nil, err
	}

	rows, err := conn.QueryContext(context.Background(), `
		SELECT DISTINCT a.link
		FROM articles a
		JOIN article_keywords ak ON ak.article_id = a.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make(map[string]bool)
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			return nil, err
		}
		links[link] = true
	}

	return links, rows.Err()
}

const selectArticlesBase = `
	SELECT
		a.title,
		a.link,
		a.description,
		s.name AS source,
		s.political_leaning,
		t.name AS topic,
		a.published_at,
		COALESCE(array_agg(k.name) FILTER (WHERE k.name IS NOT NULL), '{}') AS keywords
	FROM articles a
	JOIN sources s ON s.id = a.source_id
	LEFT JOIN topics t ON t.id = a.topic_id
	LEFT JOIN article_keywords ak ON ak.article_id = a.id
	LEFT JOIN keywords k ON k.id = ak.keyword_id
`

func GetAllArticles() ([]models.Article, error) {
	conn, err := db.Connect()
	if err != nil {
		return nil, err
	}

	rows, err := conn.QueryContext(context.Background(), selectArticlesBase+`
		GROUP BY a.id, s.name, s.political_leaning, t.name
		ORDER BY a.published_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanArticles(rows)
}

// GetArticlesByKeyword returns, most recent first, up to limit articles
// that have at least one keyword matching keyword (case-insensitive),
// across all their sources - typically used to gather the various
// articles covering the same news story before synthesizing them.
func GetArticlesByKeyword(keyword string, limit int) ([]models.Article, error) {
	conn, err := db.Connect()
	if err != nil {
		return nil, err
	}

	rows, err := conn.QueryContext(context.Background(), selectArticlesBase+`
		WHERE a.id IN (
			SELECT ak2.article_id
			FROM article_keywords ak2
			JOIN keywords k2 ON k2.id = ak2.keyword_id
			WHERE k2.name ILIKE $1
		)
		GROUP BY a.id, s.name, s.political_leaning, t.name
		ORDER BY a.published_at DESC
		LIMIT $2
	`, keyword, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanArticles(rows)
}

func scanArticles(rows *sql.Rows) ([]models.Article, error) {
	var articles []models.Article
	for rows.Next() {
		var article models.Article
		var topic sql.NullString
		if err := rows.Scan(
			&article.Title,
			&article.Link,
			&article.Description,
			&article.Source,
			&article.PoliticalLeaning,
			&topic,
			&article.PublishedAt,
			pq.Array(&article.Keywords),
		); err != nil {
			return nil, err
		}
		article.Topic = topic.String
		articles = append(articles, article)
	}

	return articles, rows.Err()
}
