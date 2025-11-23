package menu

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type menuFetchInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type menuFetchOutput struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// @Summary Fetch a menu by code
// @Tags menu
// @Param input query menuFetchInput true "input"
// @Success 200 {object} controller.Response{data=menuFetchOutput} "output"
// @Router /menu/fetch [get]
func menuFetch(ctx *gin.Context) {
	var input menuFetchInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	menuRepo := dal.NewRepo[model.TblMenu]()

	menu, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if menu == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("menu not found"))
		return
	}

	controller.Success(ctx, menuFetchOutput{
		Code: menu.Code,
		Name: menu.Name,
		Url:  menu.Url,
	})
}
