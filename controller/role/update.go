package role

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"github.com/onnttf/kit/time"
	"gorm.io/gorm"
)

type UpdateInput struct {
	Code string `json:"code" binding:"required,len=36"`
	Name string `json:"name" binding:"required,min=6,max=50"`
}

type UpdateOutput struct{}

// @Summary Update an existing role
// @Tags role
// @Param input body UpdateInput true "input"
// @Response 200 {object} controller.Response{data=UpdateOutput} "output"
// @Router /internal-api/role/update [post]
func internalApiRoleUpdate(ctx *gin.Context) {
	var input UpdateInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "role: update: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblRole]()

	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "role: update: failed, reason=query role, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		logger.Warnf(ctx, "role: update: failed, reason=role not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("role not found"))
		return
	}

	role.Name = input.Name
	role.UpdatedAt = time.NowUTC()

	if err := roleRepo.Update(ctx, database.DB, role, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		logger.Errorf(ctx, "role: update: failed, reason=update role, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "role: update: succeeded, id=%d, code=%s",
		role.Id, role.Code)

	controller.Success(ctx, UpdateOutput{})
}
