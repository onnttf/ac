package object

import (
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"

	"gorm.io/gorm"
)

type objectDeleteInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type objectDeleteOutput struct{}

// @Summary Delete an existing object
// @Tags object
// @Param input body objectDeleteInput true "input"
// @Success 200 {object} controller.Response{data=objectDeleteOutput} "output"
// @Router /api/object/delete [post]
func objectDelete(ctx *gin.Context) {
	var input objectDeleteInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	objectRepo := dal.NewRepo[model.TblObject]()

	object, err := objectRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if object == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithMsg("object not found"))
		return
	}

	object.Deleted = 1
	object.UpdatedAt = time.Now()

	if err := objectRepo.Update(ctx, database.DB, object, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, objectDeleteOutput{})
}
