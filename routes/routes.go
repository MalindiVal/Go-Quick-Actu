package routes

import (
    "github.com/gin-gonic/gin"
    "goquickactu/controllers"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()
    r.GET("/articles", articlecontrollers.GetArticles)
	r.PUT("/aggregate", articlecontrollers.AggregateArticles)
    return r
}   