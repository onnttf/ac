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

type QueryInput struct {
	Page     int    `form:"page" binding:"required,min=1" default:"1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" default:"10"`
	Email    string `json:"email" binding:"omitempty,email"`
	Name     string `json:"name" binding:"omitempty,min=1"`
}

type QueryOutput struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// @Summary Query a user by email or name
// @Tags user
// @Param input query QueryInput false "input"
// @Response 200 {object} controller.Response{data=QueryOutput} "output"
// @Router /internal-api/user/query [get]
func internalApiUserQuery(ctx *gin.Context) {
	var input QueryInput
	if err := ctx.ShouldBindQuery(&input); err != nil {
		logger.Errorf(ctx, "user: query: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		if input.Email != "" {
			return db.Where("email = ?", input.Email)
		}
		if input.Name != "" {
			return db.Where("name LIKE ?", "%"+input.Name+"%")
		}
		return db
	})
	if err != nil {
		logger.Errorf(ctx, "user: query: failed, reason=query user, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if user == nil {
		logger.Warnf(ctx, "user: query: failed, reason=user not found, input=%+v", input)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("user not found"))
		return
	}

	controller.Success(ctx, QueryOutput{
		Code:  user.Code,
		Name:  user.Name,
		Email: user.Email,
	})
}
