package user

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type userQueryInput struct {
	Page     int    `form:"page" binding:"required,min=1" default:"1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" default:"10"`
	Email    string `form:"email" binding:"omitempty,email"`
	Name     string `form:"name" binding:"omitempty,min=1"`
}

type userQueryOutput struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// @Summary Query a user by email or name
// @Tags user
// @Param input query userQueryInput false "input"
// @Response 200 {object} controller.Response{data=userQueryOutput} "output"
// @Router /user/query [get]
func userQuery(ctx *gin.Context) {
	var input userQueryInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblUser]()

	user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		if input.Email != "" {
			db.Where("email = ?", input.Email)
		}
		if input.Name != "" {
			db.Where("name LIKE ?", "%"+input.Name+"%")
		}
		return db
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if user == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("user not found"))
		return
	}

	controller.Success(ctx, userQueryOutput{
		Code:  user.Code,
		Name:  user.Name,
		Email: user.Email,
	})
}
