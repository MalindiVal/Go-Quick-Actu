package routes

import (
	"github.com/gin-gonic/gin"

	"goquickactu/controllers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/articles", controllers.GetArticles)
	r.GET("/sources", controllers.GetSources)
	r.GET("/topics", controllers.GetTopics)
	r.GET("/synthesis", controllers.GetSynthesis)
	r.POST("/aggregate", controllers.RequireAPIKey(), controllers.AggregateArticles)
	return r
}
