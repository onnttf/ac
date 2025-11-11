package role

import (
	"time"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type roleUpdateInput struct {
	Code string `json:"code" binding:"required,len=36"`
	Name string `json:"name" binding:"required,min=6,max=50"`
}

type roleUpdateOutput struct{}

// @Summary Update an existing role
// @Tags role
// @Param input body roleUpdateInput true "input"
// @Response 200 {object} controller.Response{data=roleUpdateOutput} "output"
// @Router /role/update [post]
func roleUpdate(ctx *gin.Context) {
	var input roleUpdateInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblRole]()

	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("role not found"))
		return
	}

	role.Name = input.Name
	role.UpdatedAt = time.Now()

	if err := roleRepo.Update(ctx, database.DB, role, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, roleUpdateOutput{})
}
