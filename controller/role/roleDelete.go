package role

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/time"
	"gorm.io/gorm"
)

type roleDeleteInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type roleDeleteOutput struct{}

// @Summary Delete an existing role
// @Tags role
// @Param input body roleDeleteInput true "input"
// @Response 200 {object} controller.Response{data=roleDeleteOutput} "output"
// @Router /role/delete [post]
func roleDelete(ctx *gin.Context) {
	var input roleDeleteInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblRole]()

	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("role not found"))
		return
	}

	role.Deleted = model.Deleted
	role.UpdatedAt = time.NowUTC()

	if err := roleRepo.Update(ctx, database.DB, role, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, roleDeleteOutput{})
}
