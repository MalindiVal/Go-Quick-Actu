package articledao

import (
	"context"
	"goquickactu/models"
	"goquickactu/db"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func SaveArticle(article models.Article) error {
	
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

	ctx := context.Background()
	driver, err := db.NewDriver()
	if err != nil {
		return err
	}
	defer driver.Close(ctx)
	_, err = db.ExecuteQuery(ctx, driver, query, params)


	return err
}

func GetAllArticles() models.Article[] {
	articleList := [];
	query := `
	MATCH (a:Article)-[:HAS_KEYWORD]->(k:Keyword)
	MATCH (a)-[:FROM_SOURCE]->(s:Source)
	RETURN a.title AS title, a.link AS link, a.description AS description, s.name AS source, a.PublishedAt AS publishedAt, collect(k.name) AS keywords
	ORDER BY a.PublishedAt DESC
	`
	ctx := context.Background()
	driver, err := db.NewDriver()
	if err != nil {
		// Handle error appropriately
		return nil
	}
	defer driver.Close(ctx)
	result, err := db.ExecuteQuery(ctx, driver, query, map[string]interface{}{})
	if err != nil {
		// Handle error appropriately
		return nil
	}

	if result.Next(ctx) {
		record := result.Record()
		article := models.Article{
			Title:       record.GetByIndex(0).(string),
			Link:        record.GetByIndex(1).(string),
			Description: record.GetByIndex(2).(string),
			Source:      record.GetByIndex(3).(string),
			PublishedAt: record.GetByIndex(4).(time.Time),
			Keywords:    record.GetByIndex(5).([]string),
		}
		articleList = append(articleList, article)
	}

	// Process results and return article list
	return articleList
}