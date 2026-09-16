package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"goquickactu/articledao"
	"goquickactu/ia"
	"goquickactu/models"
	"goquickactu/rss"
	"goquickactu/sources"
)

func AggregateArticles(c *gin.Context) {
	var errorsCount int

	for _, source := range sources.FeedSources {
		items, err := rss.FetchFeed(source.URL)
		if err != nil {
			log.Printf("failed to fetch feed %s: %v", source.URL, err)
			errorsCount++
			continue
		}
		log.Printf("Fetched %d items from %s", len(items), source.URL)

		for _, item := range items {
			keywords, topic, err := ia.AnalyzeContent(item.Description)
			if err != nil {
				log.Printf("failed to analyze content for %s: %v", item.Link, err)
			}

			article := rss.MapToArticle(item, source.Name)
			article.Keywords = keywords
			article.Topic = topic

			if err := articledao.SaveArticle(article); err != nil {
				log.Printf("failed to save article %s: %v", article.Link, err)
				errorsCount++
			}
		}
	}

	if errorsCount > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "articles aggregated with errors",
			"errors":  errorsCount,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "articles aggregated successfully"})
}

func GetArticles(c *gin.Context) {
	articles, err := articledao.GetAllArticles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if articles == nil {
		articles = []models.Article{}
	}

	c.JSON(http.StatusOK, articles)
}
