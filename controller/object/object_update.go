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

type objectUpdateInput struct {
	Code       string `json:"code" binding:"required,len=36"`
	Name       string `json:"name" binding:"required,min=1,max=50"`
	ParentCode string `json:"parent_code" binding:"omitempty,len=36"`
}

type objectUpdateOutput struct{}

// @Summary Update an existing object
// @Tags object
// @Param input body objectUpdateInput true "input"
// @Success 200 {object} controller.Response{data=objectUpdateOutput} "output"
// @Router /api/object/update [post]
func objectUpdate(ctx *gin.Context) {
	var input objectUpdateInput
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

	object.Name = input.Name
	if input.ParentCode != "" {
		object.ParentCode = input.ParentCode
	}
	object.UpdatedAt = time.Now()

	if err := objectRepo.Update(ctx, database.DB, object, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, objectUpdateOutput{})
}
