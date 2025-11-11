package menu

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	router := api.Group("/menu")

	router.POST("/create", menuCreate)
	router.POST("/update", menuUpdate)
	router.POST("/delete", menuDelete)
	router.GET("/fetch", menuFetch)
	router.GET("/list", menuList)
	router.GET("/query", menuQuery)
}
