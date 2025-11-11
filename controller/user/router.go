package user

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	router := api.Group("/user")

	router.POST("/create", userCreate)
	router.POST("/update", userUpdate)
	router.POST("/delete", userDelete)
	router.GET("/fetch", userFetch)
	router.GET("/list", userList)
	router.GET("/query", userQuery)
	router.POST("/role/assign", userRoleAssign)
	router.POST("/role/remove", userRoleRemove)
	router.GET("/role", userRole)
}
