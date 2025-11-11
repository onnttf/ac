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

type menuUpdateInput struct {
	Code string `json:"code" binding:"required,len=36"`
	Name string `json:"name" binding:"required,min=6,max=50"`
	Url  string `json:"url" binding:"required,url"`
}

type menuUpdateOutput struct{}

// @Summary Update an existing menu
// @Tags menu
// @Param input body menuUpdateInput true "input"
// @Response 200 {object} controller.Response{data=menuUpdateOutput} "output"
// @Router /menu/update [post]
func menuUpdate(ctx *gin.Context) {
	var input menuUpdateInput
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

	if err := validateUrl(ctx, input.Url); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	menu.Name = input.Name
	menu.Url = input.Url
	menu.UpdatedAt = time.NowUTC()

	if err := menuRepo.Update(ctx, database.DB, menu, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, menuUpdateOutput{})
}
