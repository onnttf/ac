package role

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type FetchInput struct {
	Code string `json:"code" binding:"required,len=36"`
}

type FetchOutput struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// @Summary Fetch a role by code
// @Tags role
// @Param input query FetchInput true "input"
// @Response 200 {object} controller.Response{data=FetchOutput} "output"
// @Router /internal-api/role/fetch [get]
func internalApiRoleFetch(ctx *gin.Context) {
	var input FetchInput
	if err := ctx.ShouldBind(&input); err != nil {
		logger.Errorf(ctx, "role: fetch: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblRole]()

	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", input.Code)
	})
	if err != nil {
		logger.Errorf(ctx, "role: fetch: failed, reason=query role, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		logger.Warnf(ctx, "role: fetch: failed, reason=role not found, code=%s", input.Code)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("role not found"))
		return
	}

	controller.Success(ctx, FetchOutput{
		Code: role.Code,
		Name: role.Name,
	})
}
