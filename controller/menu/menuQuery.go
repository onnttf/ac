package menu

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type menuQueryInput struct {
	Page     int    `form:"page" binding:"required,min=1" default:"1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" default:"10"`
	Url      string `form:"url" binding:"omitempty,url"`
	Name     string `form:"name" binding:"omitempty,min=1"`
}

type menuQueryOutput struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// @Summary Query menus by fields
// @Tags menu
// @Param input query menuQueryInput false "input"
// @Response 200 {object} controller.Response{data=menuQueryOutput} "output"
// @Router /menu/query [get]
func menuQuery(ctx *gin.Context) {
	var input menuQueryInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	menuRepo := dal.NewRepo[model.TblMenu]()

	menu, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		if input.Url != "" {
			db = db.Where("url = ?", input.Url)
		}
		if input.Name != "" {
			db = db.Where("name LIKE ?", "%"+input.Name+"%")
		}
		return db
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if menu == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("menu not found"))
		return
	}

	controller.Success(ctx, menuQueryOutput{
		Id:   menu.Id,
		Code: menu.Code,
		Name: menu.Name,
		Url:  menu.Url,
	})
}
