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

type DeleteInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type DeleteOutput struct{}

// @Summary Delete an existing role
// @Tags role
// @Param input body DeleteInput true "input"
// @Response 200 {object} controller.Response{data=DeleteOutput} "output"
// @Router /internal-api/role/delete [post]
func internalApiRoleDelete(ctx *gin.Context) {
	var input DeleteInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "role: delete: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblRole]()

	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "role: delete: failed, reason=query role, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		logger.Warnf(ctx, "role: delete: failed, reason=role not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("role not found"))
		return
	}

	role.Deleted = model.Deleted
	role.UpdatedAt = time.NowUTC()

	if err := roleRepo.Update(ctx, database.DB, role, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	}); err != nil {
		logger.Errorf(ctx, "role: delete: failed, reason=delete role, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	logger.Infof(ctx, "role: delete: succeeded, id=%d, code=%s",
		role.Id, role.Code)

	controller.Success(ctx, DeleteOutput{})
}
