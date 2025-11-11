package permission

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	router := api.Group("/permission")

	router.POST("/create", permissionCreate)
	router.POST("/update", permissionUpdate)
	router.POST("/delete", permissionDelete)
	router.GET("/fetch", permissionFetch)
	router.GET("/list", permissionList)
	router.GET("/query", permissionQuery)
}
