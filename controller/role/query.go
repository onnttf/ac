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

type QueryInput struct {
	Page     int    `form:"page" binding:"required,min=1" default:"1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" default:"10"`
	Name     string `json:"name" binding:"omitempty,min=1"`
}

type QueryOutput struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// @Summary Query roles by fields
// @Tags role
// @Param input query QueryInput false "input"
// @Response 200 {object} controller.Response{data=QueryOutput} "output"
// @Router /internal-api/role/query [get]
func internalApiRoleQuery(ctx *gin.Context) {
	var input QueryInput
	if err := ctx.ShouldBindQuery(&input); err != nil {
		logger.Errorf(ctx, "role: query: failed, reason=invalid input, error=%v", err)
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblRole]()

	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		if input.Name != "" {
			return db.Where("name LIKE ?", "%"+input.Name+"%")
		}
		return db
	})
	if err != nil {
		logger.Errorf(ctx, "role: query: failed, reason=query role, error=%v", err)
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		logger.Warnf(ctx, "role: query: failed, reason=role not found, input=%+v", input)
		controller.Failure(ctx, controller.ErrInvalidInput.Withmsg("role not found"))
		return
	}

	controller.Success(ctx, QueryOutput{
		Id:   role.Id,
		Code: role.Code,
		Name: role.Name,
	})
}
