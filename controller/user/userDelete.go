package user

import (
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"

	"gorm.io/gorm"
)

type userDeleteInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type userDeleteOutput struct{}

// @Summary Delete an existing user
// @Tags user
// @Param input body userDeleteInput true "input"
// @Success 200 {object} controller.Response{data=userDeleteOutput} "output"
// @Router /user/delete [post]
func userDelete(ctx *gin.Context) {
	var input userDeleteInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if user == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("user not found"))
		return
	}

	user.Deleted = model.Deleted
	user.UpdatedAt = time.Now()

	if err := userRepo.Update(ctx, database.DB, user, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userDeleteOutput{})
}
