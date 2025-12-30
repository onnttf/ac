package object

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers object-related routes.
func RegisterRoutes(api *gin.RouterGroup) {
	router := api.Group("/object")

	router.POST("/create", objectCreate)
	router.POST("/update", objectUpdate)
	router.POST("/delete", objectDelete)
	router.GET("/fetch", objectFetch)
	router.GET("/list", objectList)
	router.GET("/query", objectQuery)
}
