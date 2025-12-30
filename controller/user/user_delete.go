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
// @Router /api/user/delete [post]
func userDelete(ctx *gin.Context) {
	var input userDeleteInput
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
		"deleted":    model.Deleted,
		"updated_at": time.Now(),
	}

	userRepo := dal.NewRepo[model.TblSubject]()
	if err := userRepo.UpdateFields(ctx, database.DB, newValue, func(db *gorm.DB) *gorm.DB {
		return db.Where(condition)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userDeleteOutput{})
}
