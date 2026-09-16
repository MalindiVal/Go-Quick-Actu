package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"goquickactu/sources"
)

// GetSources returns the list of aggregated sources along with their
// indicative political leaning, so a client can display it as a legend.
func GetSources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sources": sources.FeedSources,
		"note":    "Ces orientations sont des caractérisations indicatives et simplifiées, à but de transparence, et ne constituent pas un jugement définitif.",
	})
}
