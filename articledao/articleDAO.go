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

	_, err = conn.ExecContext(context.Background(), `
		INSERT INTO articles (title, link, description, source, published_at, keywords)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (link) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			source = EXCLUDED.source,
			published_at = EXCLUDED.published_at,
			keywords = EXCLUDED.keywords
	`, article.Title, article.Link, article.Description, article.Source, article.PublishedAt, pq.Array(article.Keywords))

	return err
}

func GetAllArticles() ([]models.Article, error) {
	conn, err := db.Connect()
	if err != nil {
		return nil, err
	}

	rows, err := conn.QueryContext(context.Background(), `
		SELECT title, link, description, source, published_at, keywords
		FROM articles
		ORDER BY published_at DESC
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
			&article.PublishedAt,
			pq.Array(&article.Keywords),
		); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	return articles, rows.Err()
}
