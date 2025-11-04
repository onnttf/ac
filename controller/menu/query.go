package menu

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type QueryInput struct {
	Page     int    `form:"page" binding:"required,min=1" default:"1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" default:"10"`
	Url      string `form:"url" binding:"omitempty,url"`
	Name     string `form:"name" binding:"omitempty,min=1"`
}

type QueryOutput struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// @Summary Query menus by fields
// @Tags menu
// @Param input query QueryInput false "input"
// @Response 200 {object} controller.Response{data=QueryOutput} "output"
// @Router /internal-api/menu/query [get]
func internalApiMenuQuery(ctx *gin.Context) {
	var input QueryInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "menu: query: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	menuRepo := dal.NewRepo[model.TblMenu]()

	menu, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		if input.Url != "" {
			db.Where("url = ?", input.Url)
		}
		if input.Name != "" {
			db.Where("name LIKE ?", "%"+input.Name+"%")
		}
		return db
	})
	if err != nil {
		logger.Errorf(ctx, "menu: query: failed, reason=query menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if menu == nil {
		logger.Warnf(ctx, "menu: query: failed, reason=menu not found, input=%+v", input)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("menu not found"))
		return
	}

	controller.Success(ctx, QueryOutput{
		Id:   menu.Id,
		Code: menu.Code,
		Name: menu.Name,
		Url:  menu.Url,
	})
}
