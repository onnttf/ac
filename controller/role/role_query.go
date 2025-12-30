package role

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type roleQueryInput struct {
	Page     int    `form:"page" binding:"required,min=1" default:"1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" default:"10"`
	Name     string `form:"name" binding:"omitempty,min=1"`
}

type roleQueryOutput struct {
	Id   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// @Summary Query roles by fields
// @Tags role
// @Param input query roleQueryInput false "input"
// @Success 200 {object} controller.Response{data=roleQueryOutput} "output"
// @Router /api/role/query [get]
func roleQuery(ctx *gin.Context) {
	var input roleQueryInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	roleRepo := dal.NewRepo[model.TblSubject]()

	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		query := db.Where("type = ? AND deleted = ?", model.SubjectTypeRole, model.NotDeleted)
		if input.Name != "" {
			query = query.Where("name LIKE ?", "%"+input.Name+"%")
		}
		return query
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithMsg("role not found"))
		return
	}

	controller.Success(ctx, roleQueryOutput{
		Id:   role.Id,
		Code: role.Code,
		Name: role.Name,
	})
}
