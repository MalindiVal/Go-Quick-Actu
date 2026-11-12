package controllers

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"goquickactu/articledao"
	"goquickactu/ia"
	"goquickactu/models"
	"goquickactu/rss"
	"goquickactu/sources"
)

// aggregating is true while an aggregation runs in the background, so a
// second POST /aggregate cannot start a concurrent (and costly) run.
var aggregating atomic.Bool

// AggregateArticles starts an aggregation in the background and returns
// 202 Accepted right away: fetching every feed and analyzing each new
// article with the AI takes far longer than an HTTP request should.
// Progress and errors are written to the server logs.
//
// POST /aggregate (protected by RequireAPIKey)
func AggregateArticles(c *gin.Context) {
	if !aggregating.CompareAndSwap(false, true) {
		c.JSON(http.StatusConflict, gin.H{"error": "une agrégation est déjà en cours"})
		return
	}

	go func() {
		defer aggregating.Store(false)
		runAggregation()
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "agrégation lancée en arrière-plan"})
}

// runAggregation fetches every feed and saves its articles. Articles that
// a previous run already analyzed are skipped, so the AI is only called
// for new (or previously failed) articles.
func runAggregation() {
	start := time.Now()

	analyzed, err := articledao.GetAnalyzedLinks()
	if err != nil {
		log.Printf("aggregation aborted: failed to load analyzed articles: %v", err)
		return
	}

	var saved, skipped, analysisErrors, errorsCount int

	for _, source := range sources.FeedSources {
		items, err := rss.FetchFeed(source.URL)
		if err != nil {
			log.Printf("failed to fetch feed %s: %v", source.URL, err)
			errorsCount++
			continue
		}
		log.Printf("Fetched %d items from %s", len(items), source.URL)

		for _, item := range items {
			if analyzed[item.Link] {
				skipped++
				continue
			}

			article := rss.MapToArticle(item, source.Name)

			keywords, topic, err := ia.AnalyzeContent(item.Description)
			ok := err == nil
			if ok {
				article.Keywords = keywords
				article.Topic = topic
			} else {
				log.Printf("failed to analyze content for %s: %v", item.Link, err)
				analysisErrors++
			}

			if err := articledao.SaveArticle(article, ok); err != nil {
				log.Printf("failed to save article %s: %v", article.Link, err)
				errorsCount++
				continue
			}
			saved++
			if ok {
				analyzed[item.Link] = true
			}
		}
	}

	log.Printf("aggregation done in %s: %d saved, %d already analyzed (skipped), %d analysis errors, %d other errors",
		time.Since(start).Round(time.Second), saved, skipped, analysisErrors, errorsCount)
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
