package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"goquickactu/articledao"
	"goquickactu/ia"
)

const (
	defaultSynthesisLimit = 8
	maxSynthesisLimit     = 20
)

// GetSynthesis builds a cross-source synthesis of the news story matching
// a keyword: a neutral summary of the shared facts, plus how each source
// covering it frames the story.
//
// GET /synthesis?keyword=xxx&limit=8
func GetSynthesis(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "le paramètre 'keyword' est requis"})
		return
	}

	limit := defaultSynthesisLimit
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "le paramètre 'limit' doit être un entier positif"})
			return
		}
		limit = parsed
	}
	if limit > maxSynthesisLimit {
		limit = maxSynthesisLimit
	}

	articles, err := articledao.GetArticlesByKeyword(keyword, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(articles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "aucun article trouvé pour ce mot-clé"})
		return
	}

	synthesis, err := ia.SynthesizeArticles(articles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keyword":       keyword,
		"articles_used": articles,
		"synthesis":     synthesis,
	})
}
