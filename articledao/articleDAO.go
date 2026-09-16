package articledao

import (
	"context"

	"github.com/lib/pq"

	"goquickactu/db"
	"goquickactu/models"
)

func SaveArticle(article models.Article) error {
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

	var articleID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO articles (title, link, description, source_id, published_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (link) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			source_id = EXCLUDED.source_id,
			published_at = EXCLUDED.published_at
		RETURNING id
	`, article.Title, article.Link, article.Description, sourceID, article.PublishedAt).Scan(&articleID)
	if err != nil {
		return err
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

func GetAllArticles() ([]models.Article, error) {
	conn, err := db.Connect()
	if err != nil {
		return nil, err
	}

	rows, err := conn.QueryContext(context.Background(), `
		SELECT
			a.title,
			a.link,
			a.description,
			s.name AS source,
			s.political_leaning,
			a.published_at,
			COALESCE(array_agg(k.name) FILTER (WHERE k.name IS NOT NULL), '{}') AS keywords
		FROM articles a
		JOIN sources s ON s.id = a.source_id
		LEFT JOIN article_keywords ak ON ak.article_id = a.id
		LEFT JOIN keywords k ON k.id = ak.keyword_id
		GROUP BY a.id, s.name, s.political_leaning
		ORDER BY a.published_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		if err := rows.Scan(
			&article.Title,
			&article.Link,
			&article.Description,
			&article.Source,
			&article.PoliticalLeaning,
			&article.PublishedAt,
			pq.Array(&article.Keywords),
		); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	return articles, rows.Err()
}
