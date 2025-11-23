package user

import (
	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type userRoleInput struct {
	UserCode string `form:"user_code" binding:"required,len=36"`
}

type userRoleOutput struct {
	UserCode  string   `json:"user_code"`
	RoleCodes []string `json:"role_codes"`
}

// @Summary Query roles assigned to a user
// @Tags user
// @Param input query userRoleInput true "input"
// @Success 200 {object} controller.Response{data=userRoleOutput} "output"
// @Router /user/role [get]
func userRole(ctx *gin.Context) {
	var input userRoleInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// Validate UserCode existence
	userRepo := dal.NewRepo[model.TblUser]()
	user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ? AND deleted = 0", input.UserCode)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}
	if user == nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("user not found"))
		return
	}

	roleCodes, err := casbin.GetRolesForUser(ctx, input.UserCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userRoleOutput{UserCode: input.UserCode, RoleCodes: roleCodes})
}
