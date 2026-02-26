package db

import (
	"context"
	"rss-aggregator/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func NewDriver(uri, user, password string) (neo4j.DriverWithContext, error) {
	return neo4j.NewDriverWithContext(uri,
		neo4j.BasicAuth(user, password, ""))
}

func SaveArticle(ctx context.Context, driver neo4j.DriverWithContext,
	article models.Article) error {

	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {

			query := `
			MERGE (a:Article {link: $link})
			SET a.title = $title
			SET a.description = $description
			SET a.PublishedAt = $publishedAt
			WITH a
			UNWIND $keywords AS kw
			MERGE (k:Keyword {name: kw})
			MERGE (a)-[:HAS_KEYWORD]->(k)
			MERGE (s:Source {name: $source})
			MERGE (a)-[:FROM_SOURCE]->(s)
			`

			params := map[string]interface{}{
				"title":       article.Title,
				"link":        article.Link,
				"keywords":    article.Keywords,
				"description": article.Description,
				"publishedAt": article.PublishedAt,
			}

			return tx.Run(ctx, query, params)
		})

	return err
}
