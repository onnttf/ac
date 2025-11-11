package user

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/time"
	"gorm.io/gorm"
)

type userUpdateInput struct {
	Code  string `json:"code" binding:"required,len=36"`
	Name  string `json:"name" binding:"required,min=6,max=50" example:"Alice"`
	Email string `json:"email" binding:"required,email" example:"alice@example.com"`
}

type userUpdateOutput struct{}

// @Summary Update an existing user
// @Tags user
// @Param input body userUpdateInput true "input"
// @Response 200 {object} controller.Response{data=userUpdateOutput} "output"
// @Router /user/update [post]
func userUpdate(ctx *gin.Context) {
	var input userUpdateInput
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

	emailCount, err := userRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Unscoped().Where("email = ? AND code != ?", input.Email, input.Code)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if emailCount > 0 {

		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("email already exists"))
		return
	}

	user.Name = input.Name
	user.Email = input.Email
	user.UpdatedAt = time.NowUTC()

	if err := userRepo.Update(ctx, database.DB, user, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userUpdateOutput{})
}
