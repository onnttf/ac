package user

import (
	"github.com/gin-gonic/gin"
)

func RegisterInternalRoutes(internalApi *gin.RouterGroup) {
	router := internalApi.Group("/user")

	router.POST("/create", internalApiUserCreate)
	router.POST("/update", internalApiUserUpdate)
	router.POST("/delete", internalApiUserDelete)
	router.GET("/fetch", internalApiUserFetch)
	router.GET("/list", internalApiUserList)
	router.GET("/query", internalApiUserQuery)
}
