package user

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group) {
	g.GET("/query", queryFunc)

	g.POST("/add", addFunc)
	g.POST("/update", updateFunc)
	g.POST("/delete", deleteFunc)

	g.GET("/permission/get", permissionGetFunc)
	g.GET("/role/get", roleGetFunc)
}
