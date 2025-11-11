package user

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type userFetchInput struct {
	Code string `form:"code" binding:"required,len=36"`
}

type userFetchOutput struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// @Summary Fetch a user by code
// @Tags user
// @Param input query userFetchInput true "input"
// @Response 200 {object} controller.Response{data=userFetchOutput} "output"
// @Router /user/fetch [get]
func userFetch(ctx *gin.Context) {
	var input userFetchInput
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

	controller.Success(ctx, userFetchOutput{
		Code:  user.Code,
		Name:  user.Name,
		Email: user.Email,
	})
}
