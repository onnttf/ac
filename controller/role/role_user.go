package role

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type roleUserInput struct {
	RoleCode string `form:"role_code" binding:"required,len=36"`
}

type roleUserOutput struct {
	RoleCode  string   `json:"role_code"`
	UserCodes []string `json:"user_codes"`
}

// @Summary Query users assigned to a role
// @Tags role
// @Param input query roleUserInput true "input"
// @Success 200 {object} controller.Response{data=roleUserOutput} "output"
// @Router /api/role/user [get]
func roleUser(ctx *gin.Context) {
	var input roleUserInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Validate RoleCode existence
	roleRepo := dal.NewRepo[model.TblSubject]()
	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ? AND type = ? AND deleted = ?", input.RoleCode, model.SubjectTypeRole, model.NotDeleted)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("role not found"))
		return
	}

	userCodes, err := casbin.GetUsersForRole(ctx, input.RoleCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, roleUserOutput{
		RoleCode:  input.RoleCode,
		UserCodes: userCodes,
	})
}
