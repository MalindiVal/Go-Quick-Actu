package articlecontrollers

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "goquickactu/models"
	"goquickactu/articledao"
	"goquickactu/rss"
	"goquickactu/ia"
)

CONST(
	urls := []string{
		"https://www.franceinfo.fr/france.rss",
		"https://www.afp.com/fr/actus/afp_communique/all/feed",
		"https://www.lefigaro.fr/rss/figaro_actualites.xml",
	}
)

func AggregateArticles(c *gin.Context) {
	// Fetch and process articles from RSS feeds
	for _, url := range urls {
		items, _ := rss.FetchFeed(url)
		log.Printf("Fetched %d items from %s", len(items), url)
		for _, item := range items {
			keywords, _ := ia.ExtractKeywords(item.Description)
			article := rss.MapToArticle(item, "example-source")
			article.Keywords = keywords
			articledao.SaveArticle(article)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Articles aggregated successfully"})
	
}

func GetArticles(c *gin.Context) {
	var articles []models.Article
	
	articles = articledao.GetAllArticles()

	c.JSON(http.StatusOK, articles)
	
}   