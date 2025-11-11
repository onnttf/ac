package menu

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/time"
	"gorm.io/gorm"
)

type menuDeleteInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type menuDeleteOutput struct{}

// @Summary Delete an existing menu
// @Tags menu
// @Param input body menuDeleteInput true "input"
// @Response 200 {object} controller.Response{data=menuDeleteOutput} "output"
// @Router /menu/delete [post]
func menuDelete(ctx *gin.Context) {
	var input menuDeleteInput
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

	menu.Deleted = model.Deleted
	menu.UpdatedAt = time.NowUTC()

	if err := menuRepo.Update(ctx, database.DB, menu, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, menuDeleteOutput{})
}
