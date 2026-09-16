package routes

import (
	"github.com/gin-gonic/gin"

	"goquickactu/controllers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/articles", controllers.GetArticles)
	r.GET("/sources", controllers.GetSources)
	r.POST("/aggregate", controllers.AggregateArticles)
	return r
}
