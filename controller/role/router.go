package role

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	router := r.Group("/role")

	router.POST("/create", internalApiRoleCreate)
	router.POST("/update", internalApiRoleUpdate)
	router.POST("/delete", internalApiRoleDelete)
	router.GET("/fetch", internalApiRoleFetch)
	router.GET("/list", internalApiRoleList)
	router.GET("/query", internalApiRoleQuery)
}
