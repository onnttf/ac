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
// @Response 200 {object} controller.Response{data=roleUserOutput} "output"
// @Router /role/user [get]
func roleUser(ctx *gin.Context) {
	var input roleUserInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Validate RoleCode existence
	roleRepo := dal.NewRepo[model.TblRole]()
	role, err := roleRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ? AND deleted = 0", input.RoleCode)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if role == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("role not found"))
		return
	}

	// Retrieve users assigned to the role
	userCodes, err := casbin.GetRoleUsers(ctx, input.RoleCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, roleUserOutput{
		RoleCode:  input.RoleCode,
		UserCodes: userCodes,
	})
}
