package user

import (
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/util"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"

	"gorm.io/gorm"
)

type userCreateInput struct {
	Name string `json:"name" binding:"required,min=6,max=50" example:"Alice"`
}

type userCreateOutput struct {
	Code string `json:"code"`
}

// @Summary Create a new user
// @Tags user
// @Param input body userCreateInput true "input"
// @Success 200 {object} controller.Response{data=userCreateOutput} "output"
// @Router /user/create [post]
func userCreate(ctx *gin.Context) {
	var input userCreateInput
	if err := ctx.ShouldBind(&input); err != nil {

		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	userRepo := dal.NewRepo[model.TblSubject]()

	emailCount, err := userRepo.Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Unscoped().Where("type = ? AND name = ?", model.SubjectTypeUser, input.Name)
	})
	if err != nil {

		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if emailCount > 0 {

		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("user name already exists"))
		return
	}

	// Create new user
	now := time.Now()
	newValue := &model.TblSubject{
		Type:      model.SubjectTypeUser,
		Code:      util.GenerateCode(),
		Name:      input.Name,
		Status:    model.StatusEnabled.Int64(),
		Deleted:   model.NotDeleted,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := userRepo.Insert(ctx, database.DB, newValue); err != nil {

		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userCreateOutput{Code: newValue.Code})
}
