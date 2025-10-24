package user

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
	Code string `form:"code" binding:"required,len=36"`
}

type FetchOutput struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// @Summary Fetch a user by code
// @Tags user
// @Param input query FetchInput true "input"
// @Response 200 {object} controller.Response{data=FetchOutput} "output"
// @Router /internal-api/user/fetch [get]
func internalApiUserFetch(ctx *gin.Context) {
	var input FetchInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "user: fetch: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "user: fetch: failed, reason=query user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if user == nil {
		logger.Warnf(ctx, "user: fetch: failed, reason=user not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("user not found"))
		return
	}

	controller.Success(ctx, FetchOutput{
		Code:  user.Code,
		Name:  user.Name,
		Email: user.Email,
	})
}
