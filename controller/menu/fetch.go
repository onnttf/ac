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

type FetchInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type FetchOutput struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// @Summary Fetch a menu by code
// @Tags menu
// @Param input query FetchInput true "input"
// @Response 200 {object} controller.Response{data=FetchOutput} "output"
// @Router /internal-api/menu/fetch [get]
func internalApiMenuFetch(ctx *gin.Context) {
	var input FetchInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "menu: fetch: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	menuRepo := dal.NewRepo[model.TblMenu]()

	menu, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "menu: fetch: failed, reason=query menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if menu == nil {
		logger.Warnf(ctx, "menu: fetch: failed, reason=menu not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("menu not found"))
		return
	}

	controller.Success(ctx, FetchOutput{
		Code: menu.Code,
		Name: menu.Name,
		Url:  menu.Url,
	})
}
