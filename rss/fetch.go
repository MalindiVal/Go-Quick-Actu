package rss

import (
	"time"

	"rss-aggregator/models"

	"github.com/mmcdole/gofeed"
)

func MapToArticle(item *gofeed.Item, source string) models.Article {
	published := time.Now()
	if item.PublishedParsed != nil {
		published = *item.PublishedParsed
	}

	return models.Article{
		Title:       item.Title,
		Link:        item.Link,
		Description: item.Description,
		Source:      source,
		PublishedAt: published,
	}
}
