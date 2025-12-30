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

type userUpdateInput struct {
	Code string `json:"code" binding:"required,len=36"`
	Name string `json:"name" binding:"required,min=6,max=50" example:"Alice"`
}

type userUpdateOutput struct{}

// @Summary Update an existing user
// @Tags user
// @Param input body userUpdateInput true "input"
// @Success 200 {object} controller.Response{data=userUpdateOutput} "output"
// @Router /api/user/update [post]
func userUpdate(ctx *gin.Context) {
	var input userUpdateInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	condition := map[string]any{
		"type":    model.SubjectTypeUser,
		"code":    input.Code,
		"deleted": model.NotDeleted,
	}

	newValue := map[string]any{
		"name":       input.Name,
		"updated_at": time.Now(),
	}

	userRepo := dal.NewRepo[model.TblSubject]()
	if err := userRepo.UpdateFields(ctx, database.DB, newValue, func(db *gorm.DB) *gorm.DB {
		return db.Where(condition)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userUpdateOutput{})
}
