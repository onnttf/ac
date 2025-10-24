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

type UpdateInput struct {
	Code string `json:"code" binding:"required,len=36"`
	Name string `json:"name" binding:"required,min=6,max=50"`
	Url  string `json:"url" binding:"required,url"`
}

type UpdateOutput struct{}

// @Summary Update an existing menu
// @Tags menu
// @Param input body UpdateInput true "input"
// @Response 200 {object} controller.Response{data=UpdateOutput} "output"
// @Router /internal-api/menu/update [post]
func internalApiMenuUpdate(ctx *gin.Context) {
	var input UpdateInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "menu: update: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	menuRepo := dal.NewRepo[model.TblMenu]()

	menu, err := menuRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "menu: update: failed, reason=query menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if menu == nil {
		logger.Warnf(ctx, "menu: update: failed, reason=menu not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("menu not found"))
		return
	}

	if err := validateUrl(ctx, input.Url); err != nil {
		logger.Errorf(ctx, "menu: update: failed, reason=check url, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	menu.Name = input.Name
	menu.Url = input.Url
	menu.UpdatedAt = time.NowUTC()

	if err := menuRepo.Update(ctx, database.DB, menu, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		logger.Errorf(ctx, "menu: update: failed, reason=update menu, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "menu: update: succeeded, id=%d, code=%s, url=%s",
		menu.Id, menu.Code, menu.Url)

	controller.Success(ctx, UpdateOutput{})
}
