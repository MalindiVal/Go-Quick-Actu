package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"goquickactu/topics"
)

// GetTopics returns the fixed taxonomy of news topics articles are
// classified into.
func GetTopics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"topics": topics.Topics})
}
