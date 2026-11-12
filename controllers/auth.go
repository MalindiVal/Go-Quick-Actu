package controllers

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// RequireAPIKey protects a route with the key set in the ADMIN_API_KEY
// environment variable, sent by the client in the "X-API-Key" header.
//
// If ADMIN_API_KEY is not set, the route is refused for everyone rather
// than left open: it triggers paid AI calls.
func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := os.Getenv("ADMIN_API_KEY")
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "route désactivée : ADMIN_API_KEY n'est pas configurée"})
			return
		}

		given := c.GetHeader("X-API-Key")
		if subtle.ConstantTimeCompare([]byte(given), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "clé API manquante ou invalide"})
			return
		}

		c.Next()
	}
}
