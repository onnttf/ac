package menu

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	router := r.Group("/menu")

	router.POST("/create", internalApiMenuCreate)
	router.POST("/update", internalApiMenuUpdate)
	router.POST("/delete", internalApiMenuDelete)
	router.GET("/fetch", internalApiMenuFetch)
	router.GET("/list", internalApiMenuList)
	router.GET("/query", internalApiMenuQuery)
}
