package user

import (
	"strings"

	"ac/bootstrap/database"
	"ac/controller"
	"ac/model"
	"ac/service/casbin"

	"github.com/gin-gonic/gin"
	"github.com/onnttf/kit/dal"
	"gorm.io/gorm"
)

type userRoleRemoveInput struct {
	UserCode  string   `json:"user_code" binding:"required,len=36"`
	RoleCodes []string `json:"role_codes" binding:"required,min=1,dive,len=36"`
}

type userRoleRemoveOutput struct {
	RemovedCount int `json:"removed_count"`
}

// @Summary Remove roles from a user
// @Tags user
// @Param input body userRoleRemoveInput true "input"
// @Response 200 {object} controller.Response{data=userRoleRemoveOutput} "output"
// @Router /user/role/remove [post]
func userRoleRemove(ctx *gin.Context) {
	var input userRoleRemoveInput
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

	// 2. Validate roles existence and ensure roles are assigned to the user
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

	// Check for invalid roles
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

	// 3. Ensure roles are assigned to the user before attempting removal
	assignedRoles, err := casbin.GetUserRoles(ctx, input.UserCode)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	// Convert assignedRoles to a set for quick lookup
	assignedRolesSet := make(map[string]struct{})
	for _, role := range assignedRoles {
		assignedRolesSet[role] = struct{}{}
	}

	var rolesToRemove []string
	for _, roleCode := range input.RoleCodes {
		if _, assigned := assignedRolesSet[roleCode]; assigned {
			rolesToRemove = append(rolesToRemove, roleCode)
		}
	}

	if len(rolesToRemove) == 0 {
		controller.Failure(ctx, controller.ErrInvalidInput.WithHint("no matching roles assigned to user"))
		return
	}

	// 4. Remove roles from the user
	removedCount, err := casbin.RemoveRolesFromUser(ctx, input.UserCode, rolesToRemove)
	if err != nil {
		controller.Failure(ctx, controller.ErrSystemError.WithError(err))
		return
	}

	controller.Success(ctx, userRoleRemoveOutput{RemovedCount: removedCount})
}
