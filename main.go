package main

import (
	"context"
	"log"
	"rss-aggregator/db"
	"rss-aggregator/ia"
	"rss-aggregator/rss"

	"github.com/gin-gonic/gin"
)

func main() {
	uri := "bolt://localhost:7687"
	username := "neo4j"
	password := "Acturoot"

	driver, err := db.NewDriver(uri, username, password)
	if err != nil {
		log.Fatal("Neo4j connection error:", err)
	}
	defer driver.Close(context.Background())

	r := gin.Default()

	r.POST("/aggregate", func(c *gin.Context) {

		urls := []string{
			"https://www.franceinfo.fr/france.rss",
			"https://www.afp.com/fr/actus/afp_communique/all/feed",
			"https://www.lefigaro.fr/rss/figaro_actualites.xml",
			"https://www.lemonde.fr/rss/une.xml",
			"https://www.nouvelobs.com/rss.xml",
			"https://www.huffingtonpost.fr/feeds/index.xml",
			"https://www.europe1.fr/rss.xml",
		}

		for _, url := range urls {
			items, _ := rss.FetchFeed(url)
			log.Printf("Fetched %d items from %s", len(items), url)

			for _, item := range items {
				keywords, _ := ia.ExtractKeywords(item.Description)
				article := rss.MapToArticle(item, "example-source")
				article.Keywords = keywords
				db.SaveArticle(context.Background(), driver, article)
			}
		}

		c.JSON(200, gin.H{"status": "done"})
	})

	r.Run(":8080")
}
