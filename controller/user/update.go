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

type UpdateInput struct {
	Code  string `json:"code" binding:"required,len=36"`
	Name  string `json:"name" binding:"required,min=6,max=50" example:"Alice"`
	Email string `json:"email" binding:"required,email" example:"alice@example.com"`
}

type UpdateOutput struct{}

// @Summary Update an existing user
// @Tags user
// @Param input body UpdateInput true "input"
// @Response 200 {object} controller.Response{data=UpdateOutput} "output"
// @Router /internal-api/user/update [post]
func internalApiUserUpdate(ctx *gin.Context) {
	var input UpdateInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "user: update: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "user: update: failed, reason=query user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if user == nil {
		logger.Warnf(ctx, "user: update: failed, reason=user not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("user not found"))
		return
	}

	emailCount, err := userRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Unscoped().Where("email = ? AND code != ?", input.Email, input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "user: update: failed, reason=query email, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if emailCount > 0 {
		logger.Warnf(ctx, "user: update: failed, reason=email already exists, email=%s", input.Email)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("email already exists"))
		return
	}

	user.Name = input.Name
	user.Email = input.Email
	user.UpdatedAt = time.NowUTC()

	if err := userRepo.Update(ctx, database.DB, user, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		logger.Errorf(ctx, "user: update: failed, reason=update user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "user: update: succeeded, id=%d, code=%s, email=%s",
		user.Id, user.Code, user.Email)

	controller.Success(ctx, UpdateOutput{})
}
