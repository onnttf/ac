package menu

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/time"
	"gorm.io/gorm"
)

type DeleteInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type DeleteOutput struct{}

// @Summary Delete an existing menu
// @Tags menu
// @Param input body DeleteInput true "input"
// @Response 200 {object} controller.Response{data=DeleteOutput} "output"
// @Router /internal-api/menu/delete [post]
func internalApiMenuDelete(ctx *gin.Context) {
	var input DeleteInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "menu: delete: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	menuRepo := dal.NewRepo[model.TblRole]()

	menu, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "menu: delete: failed, reason=query menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if menu == nil {
		logger.Warnf(ctx, "menu: delete: failed, reason=menu not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("menu not found"))
		return
	}

	menu.Deleted = model.Deleted
	menu.UpdatedAt = time.NowUTC()

	if err := menuRepo.Update(ctx, database.DB, menu, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		logger.Errorf(ctx, "menu: delete: failed, reason=delete menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "menu: delete: succeeded, id=%d, code=%s",
		menu.Id, menu.Code)

	controller.Success(ctx, DeleteOutput{})
}
