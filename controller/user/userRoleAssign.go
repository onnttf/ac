package user

import (
	"slices"
	"strings"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type userRoleAssignInput struct {
	UserCode  string   `json:"user_code" binding:"required,len=36"`
	RoleCodes []string `json:"role_codes" binding:"required,min=1,dive,len=36"`
}

type userRoleAssignOutput struct{}

// @Summary Assign roles to a user
// @Tags user
// @Param input body userRoleAssignInput true "input"
// @Success 200 {object} controller.Response{data=userRoleAssignOutput} "output"
// @Router /user/role/assign [post]
func userRoleAssign(ctx *gin.Context) {
	var input userRoleAssignInput
	if err := ctx.ShouldBind(&input); err != nil {
		controller.Failure(ctx, controller.ErrInvalidInput.WithError(err))
		return
	}

	// 1. Validate user existence
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

	// 2. Check if user already has the roles
	userRoles, err := casbin.GetRolesForUser(ctx, input.UserCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// Ensure no role is already assigned to the user
	for _, roleCode := range input.RoleCodes {
		if slices.Contains(userRoles, roleCode) {
			controller.Failure(ctx, controller.ErrInvalidInput.WithHint("role already assigned to the user: "+roleCode))
			return
		}
	}

	// 3. Validate roles existence
	roleRepo := dal.NewRepo[model.TblRole]()
	roles, err := roleRepo.Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code IN ? AND deleted = 0", input.RoleCodes)
	})
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// Collect valid roles
	validRoles := make(map[string]struct{})
	for _, role := range roles {
		validRoles[role.Code] = struct{}{}
	}

	// Check if there are any invalid roles
	var invalidRoles []string
	for _, roleCode := range input.RoleCodes {
		if _, exists := validRoles[roleCode]; !exists {
			invalidRoles = append(invalidRoles, roleCode)
		}
	}

	if len(invalidRoles) > 0 {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("invalid roles: "+strings.Join(invalidRoles, ",")))
		return
	}

	// 4. Assign roles to the user
	if err := casbin.AssignRolesToUser(ctx, input.UserCode, input.RoleCodes); err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userRoleAssignOutput{})
}
