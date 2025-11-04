package user

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

// @Summary Delete an existing user
// @Tags user
// @Param input body DeleteInput true "input"
// @Response 200 {object} controller.Response{data=DeleteOutput} "output"
// @Router /internal-api/user/delete [post]
func internalApiUserDelete(ctx *gin.Context) {
	var input DeleteInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "user: delete: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "user: delete: failed, reason=query user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if user == nil {
		logger.Warnf(ctx, "user: delete: failed, reason=user not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("user not found"))
		return
	}

	user.Deleted = model.Deleted
	user.UpdatedAt = time.NowUTC()

	if err := userRepo.Update(ctx, database.DB, user, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		logger.Errorf(ctx, "user: delete: failed, reason=delete user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "user: delete: succeeded, id=%d, code=%s",
		user.Id, user.Code)

	controller.Success(ctx, DeleteOutput{})
}
