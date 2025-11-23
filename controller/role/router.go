package role

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers role-related routes.
func RegisterRoutes(api *gin.RouterGroup) {
	router := api.Group("/role")

	router.POST("/create", roleCreate)
	router.POST("/update", roleUpdate)
	router.POST("/delete", roleDelete)
	router.GET("/fetch", roleFetch)
	router.GET("/list", roleList)
	router.GET("/query", roleQuery)

	router.POST("/user/assign", roleUserAssign)
	router.POST("/user/remove", roleUserRemove)
	router.GET("/user", roleUser)
}
